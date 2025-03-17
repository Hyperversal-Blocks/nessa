package api

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"keyless-auth/repository/session"
	"keyless-auth/repository/user"
	"keyless-auth/signer"
)

// [WIP]
var (
	ERR_CHALLENGE_EXISTS     = errors.New("challenge already exists")
	ERR_CHALLENGE_NOT_FOUND  = errors.New("challenge not found")
	ERR_USER_SESSION_EXISTS  = errors.New("user session already exists")
	ERR_INVALID_PUBLIC_KEY   = errors.New("invalid public key")
	ERR_VALIDATING_CHALLENGE = errors.New("err while validating challenge")
)

type VerifyChallengeRequest struct {
	DecryptedChallenge string `json:"decrypted_challenge"`
	Signature          string `json:"signature"`
	EphemeralPK        string `json:"ephemeral_pk"`
	ChallengerPK       string `json:"challenger_pk"`
}

type VerifyChallengeResponse struct {
	SessionToken string `json:"session_token"`
	PublicKey    string `json:"public_key"`
}

type ephemeral struct {
	SignerObject    signer.Signer
	UserPubKeyStore user.Repo
	SessionStore    session.Store
}

type ChallengeReq struct {
	PubKey string `json:"public_key"`
}

type ChallengePayloadResp struct {
	Challenge       string `json:"challenge,"`
	HashedSharedKey string `json:"hashed_signature,"`
	EphemeralPubKey string `json:"ephemeral_pub_key"`
	Signature       string `json:"signature"`
}

func (e *ephemeral) Metamask() string {
	return ""
}

type Ephemeral interface {
	VerifyChallengeHandler(w http.ResponseWriter, r *http.Request)
	ChallengeHandler(w http.ResponseWriter, r *http.Request)
}

func NewEphemeralHandler(s signer.Signer, u user.Repo, store session.Store) Ephemeral {
	return &ephemeral{
		SignerObject:    s,
		UserPubKeyStore: u,
		SessionStore:    store,
	}
}

func (e *ephemeral) VerifyChallengeHandler(w http.ResponseWriter, r *http.Request) {
	var req VerifyChallengeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	userPK, err := e.SignerObject.PublicKeyFromBytes(req.ChallengerPK)
	if err != nil {
		http.Error(w, "Invalid public key", http.StatusBadRequest)
		return
	}

	// TODO: reset the challenge and delete the keys after debugging
	challenge, err := e.SessionStore.GetActiveChallengeForUser(r.Context(), req.ChallengerPK)
	if err != nil {
		http.Error(w, "Invalid challenge", http.StatusBadRequest)
		return
	}

	ephemeralSigner, err := signer.NewFromKey(challenge.EphemeralPrivKey)
	if err != nil {
		http.Error(w, "unable to parse keys", http.StatusInternalServerError)
		return
	}

	sharedKey := ephemeralSigner.GetSharedKey(*userPK)

	decryptedMessage, _ := ephemeralSigner.DecryptMessage(sharedKey, challenge.CipheredChallenge)

	if req.DecryptedChallenge != decryptedMessage {
		http.Error(w, "challenge failed", http.StatusBadRequest)
		return
	}

	userSigner, err := signer.New()
	if err != nil {
		http.Error(w, "unable to register user", http.StatusInternalServerError)
		return
	}

	err = e.UserPubKeyStore.SavePubKeyUser(r.Context(), &user.UserWithPubKey{
		PubKey:              req.ChallengerPK,
		ID:                  uuid.New().String(),
		EncryptedPrivateKey: hex.EncodeToString(crypto.FromECDSA(userSigner.GetPrivateKey())),
	})
	if err != nil {
		http.Error(w, "Failed to save user", http.StatusBadRequest)
		return
	}
	json.NewEncoder(w).Encode(VerifyChallengeResponse{
		SessionToken: "verified",
	})
}

func (e *ephemeral) ChallengeHandler(w http.ResponseWriter, r *http.Request) {
	var req ChallengeReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if len(strings.TrimSpace(req.PubKey)) == 0 {
		http.Error(w, "invalid public key", http.StatusBadRequest)
		return
	}

	existingChallenge, err := e.challengeValidator(r.Context(), req.PubKey)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if existingChallenge != nil {
		err = e.SessionStore.DiscardExistingChallenge(r.Context(), req.PubKey)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
	}

	userPK, err := e.SignerObject.PublicKeyFromBytes(req.PubKey)
	if err != nil {
		http.Error(w, "invalid or unsupported public key", http.StatusBadRequest)
		return
	}

	randomSigner, err := signer.New()
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	sharedKey := randomSigner.GetSharedKey(*userPK)

	hashedKey, ciphered, err := randomSigner.EncryptAndGetChallengeHash(sharedKey, uuid.NewString())
	if err != nil {
		http.Error(w, "unable to send challenge", http.StatusInternalServerError)
		return
	}

	serverSig, err := e.SignerObject.Sign(hex.EncodeToString(crypto.FromECDSAPub(randomSigner.GetPublicKey())))
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	err = e.SessionStore.SaveChallenge(r.Context(), &session.ChallengeRecord{
		ServerSignature:   serverSig,
		UserPubKey:        req.PubKey,
		EphemeralPubKey:   hex.EncodeToString(crypto.FromECDSAPub(randomSigner.GetPublicKey())),
		EphemeralPrivKey:  hex.EncodeToString(crypto.FromECDSA(randomSigner.GetPrivateKey())),
		HashedSharedKey:   hashedKey,
		CipheredChallenge: ciphered,
		CreatedAt:         time.Now().UTC().Unix(),
		TTL:               86400,
	})

	json.NewEncoder(w).Encode(ChallengePayloadResp{
		Challenge:       ciphered,
		HashedSharedKey: hashedKey,
		EphemeralPubKey: hex.EncodeToString(crypto.FromECDSAPub(randomSigner.GetPublicKey())),
		Signature:       serverSig,
	})
}

func (e *ephemeral) challengeValidator(ctx context.Context, pubKey string) (*session.ChallengeRecord, error) {
	userSessionExists, err := e.SessionStore.UserExistsForSession(ctx, pubKey)
	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, fmt.Errorf("get existing user session: %w: %w", ERR_VALIDATING_CHALLENGE, err)
	}

	// Fetch the latest active challenge for the user
	challengeRec, err := e.SessionStore.GetActiveChallengeForUser(ctx, pubKey)
	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, fmt.Errorf("get active challenge: %w: %w", ERR_VALIDATING_CHALLENGE, err)
	}

	if !userSessionExists || challengeRec == nil {
		return nil, nil
	}

	ephemeralKeyExists, err := e.SessionStore.EphemeralKeyExists(ctx, pubKey, challengeRec.EphemeralPubKey)
	if err != nil {
		return nil, fmt.Errorf("ephemeral: %w: %w", ERR_VALIDATING_CHALLENGE, err)
	}

	if !ephemeralKeyExists {
		return nil, nil
	}

	return challengeRec, nil
}
