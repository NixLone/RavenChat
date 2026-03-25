package transport

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"ravenchat/apps/api/internal/audit"
	"ravenchat/apps/api/internal/auth"
	"ravenchat/apps/api/internal/boards"
	"ravenchat/apps/api/internal/chats"
	"ravenchat/apps/api/internal/files"
	"ravenchat/apps/api/internal/groups"
	"ravenchat/apps/api/internal/presence"
	"ravenchat/apps/api/internal/users"
)

type Server struct {
	authProvider auth.Provider
	jwt          *auth.JWTManager
	users        *users.PGRepository
	chats        *chats.Service
	groups       *groups.Repository
	boards       *boards.Repository
	files        *files.Service
	audit        *audit.Service
	presence     *presence.Service
	wsHub        *Hub
}

func NewServer(authProvider auth.Provider, jwt *auth.JWTManager, users *users.PGRepository, chats *chats.Service, groups *groups.Repository, boards *boards.Repository, files *files.Service, audit *audit.Service, presence *presence.Service, wsHub *Hub) *Server {
	return &Server{authProvider: authProvider, jwt: jwt, users: users, chats: chats, groups: groups, boards: boards, files: files, audit: audit, presence: presence, wsHub: wsHub}
}

func (s *Server) Router(frontendURL string) http.Handler {
	r := chi.NewRouter()
	r.Use(cors.Handler(cors.Options{AllowedOrigins: []string{frontendURL, "http://localhost:5173"}, AllowedMethods: []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"}, AllowedHeaders: []string{"Accept", "Authorization", "Content-Type"}, AllowCredentials: true}))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })

	r.Post("/api/v1/auth/login", s.handleLogin)

	r.Group(func(pr chi.Router) {
		pr.Use(s.authMiddleware)
		pr.Get("/api/v1/auth/me", s.handleMe)
		pr.Get("/api/v1/users", s.handleUsers)
		pr.Get("/api/v1/chats", s.handleChats)
		pr.Get("/api/v1/chats/{chatID}/messages", s.handleMessages)
		pr.Post("/api/v1/chats/{chatID}/messages", s.handleSendMessage)
		pr.Post("/api/v1/chats/{chatID}/ws", s.handleChatWS)
		pr.Get("/api/v1/groups", s.handleGroups)
		pr.Get("/api/v1/groups/{groupID}/members", s.handleGroupMembers)
		pr.Post("/api/v1/groups/{groupID}/members", s.handleGroupMemberAdd)
		pr.Delete("/api/v1/groups/{groupID}/members/{userID}", s.handleGroupMemberRemove)
		pr.Get("/api/v1/boards/{boardID}", s.handleBoard)
		pr.Post("/api/v1/boards/{boardID}/tasks", s.handleCreateTask)
		pr.Patch("/api/v1/tasks/{taskID}/move", s.handleMoveTask)
		pr.Post("/api/v1/files/upload", s.handleUpload)
		pr.Get("/api/v1/audit/events", s.handleAudit)
	})
	return r
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	identity, err := s.authProvider.Authenticate(r.Context(), req.Login, req.Password)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	token, _ := s.jwt.Mint(identity)
	jsonResp(w, map[string]any{"token": token, "user": identity})
}

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	identity := mustIdentity(r.Context())
	u, err := s.users.ByID(r.Context(), identity.UserID)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	p, _ := s.users.Profile(r.Context(), identity.UserID)
	_ = s.presence.MarkOnline(r.Context(), identity.UserID)
	jsonResp(w, map[string]any{"user": u, "profile": p})
}

func (s *Server) handleUsers(w http.ResponseWriter, r *http.Request) {
	items, err := s.users.List(r.Context())
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	jsonResp(w, items)
}

func (s *Server) handleChats(w http.ResponseWriter, r *http.Request) {
	identity := mustIdentity(r.Context())
	items, err := s.chats.ListForUser(r.Context(), identity.UserID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	jsonResp(w, items)
}

func (s *Server) handleMessages(w http.ResponseWriter, r *http.Request) {
	chatID, _ := uuid.Parse(chi.URLParam(r, "chatID"))
	items, err := s.chats.Messages(r.Context(), chatID, 50)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	jsonResp(w, items)
}

func (s *Server) handleSendMessage(w http.ResponseWriter, r *http.Request) {
	identity := mustIdentity(r.Context())
	chatID, _ := uuid.Parse(chi.URLParam(r, "chatID"))
	var req struct {
		Body   string     `json:"body"`
		FileID *uuid.UUID `json:"file_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	m, err := s.chats.SendMessage(r.Context(), chatID, identity.UserID, req.Body, req.FileID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	jsonResp(w, m)
}

func (s *Server) handleChatWS(w http.ResponseWriter, r *http.Request) {
	chatID, _ := uuid.Parse(chi.URLParam(r, "chatID"))
	s.wsHub.Handle(w, r, chatID)
}

func (s *Server) handleGroups(w http.ResponseWriter, r *http.Request) {
	identity := mustIdentity(r.Context())
	items, err := s.groups.ListForUser(r.Context(), identity.UserID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	jsonResp(w, items)
}

func (s *Server) handleGroupMembers(w http.ResponseWriter, r *http.Request) {
	groupID, _ := uuid.Parse(chi.URLParam(r, "groupID"))
	items, err := s.groups.Members(r.Context(), groupID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	jsonResp(w, items)
}

func (s *Server) handleGroupMemberAdd(w http.ResponseWriter, r *http.Request) {
	identity := mustIdentity(r.Context())
	groupID, _ := uuid.Parse(chi.URLParam(r, "groupID"))
	var req struct {
		UserID uuid.UUID `json:"user_id"`
		Role   string    `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	if req.Role == "" {
		req.Role = "group_member"
	}
	if err := s.groups.AddMember(r.Context(), groupID, req.UserID, req.Role); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	_ = s.audit.Log(r.Context(), identity.UserID, "group.member_added", "group", groupID, map[string]any{"member_user_id": req.UserID})
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleGroupMemberRemove(w http.ResponseWriter, r *http.Request) {
	identity := mustIdentity(r.Context())
	groupID, _ := uuid.Parse(chi.URLParam(r, "groupID"))
	userID, _ := uuid.Parse(chi.URLParam(r, "userID"))
	if err := s.groups.RemoveMember(r.Context(), groupID, userID); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	_ = s.audit.Log(r.Context(), identity.UserID, "group.member_removed", "group", groupID, map[string]any{"member_user_id": userID})
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleBoard(w http.ResponseWriter, r *http.Request) {
	boardID, _ := uuid.Parse(chi.URLParam(r, "boardID"))
	cols, err := s.boards.Columns(r.Context(), boardID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	tasks, err := s.boards.Tasks(r.Context(), boardID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	jsonResp(w, map[string]any{"columns": cols, "tasks": tasks})
}

func (s *Server) handleCreateTask(w http.ResponseWriter, r *http.Request) {
	identity := mustIdentity(r.Context())
	boardID, _ := uuid.Parse(chi.URLParam(r, "boardID"))
	var req boards.Task
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	req.BoardID = boardID
	req.CreatorID = identity.UserID
	if req.Status == "" {
		req.Status = "todo"
	}
	t, err := s.boards.CreateTask(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	jsonResp(w, t)
}

func (s *Server) handleMoveTask(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ColumnID uuid.UUID `json:"column_id"`
		Status   string    `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	taskID, _ := uuid.Parse(chi.URLParam(r, "taskID"))
	if err := s.boards.MoveTask(r.Context(), taskID, req.ColumnID, req.Status); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleUpload(w http.ResponseWriter, r *http.Request) {
	identity := mustIdentity(r.Context())
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	f, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	defer f.Close()
	stored, err := s.files.Upload(r.Context(), identity.UserID, f, header, header.Header.Get("Content-Type"))
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	jsonResp(w, stored)
}

func (s *Server) handleAudit(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	items, err := s.audit.List(r.Context(), limit)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	jsonResp(w, items)
}

func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authz := r.Header.Get("Authorization")
		if len(authz) < 8 {
			http.Error(w, "unauthorized", 401)
			return
		}
		token := authz[7:]
		identity, err := s.jwt.Parse(token)
		if err != nil {
			http.Error(w, "unauthorized", 401)
			return
		}
		ctx := context.WithValue(r.Context(), identityCtxKey{}, identity)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type identityCtxKey struct{}

func mustIdentity(ctx context.Context) auth.Identity {
	identity, ok := ctx.Value(identityCtxKey{}).(auth.Identity)
	if !ok {
		panic(errors.New("missing identity"))
	}
	return identity
}

func jsonResp(w http.ResponseWriter, payload any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(payload)
}

func NewRedis(addr, password string, dbNum int) *redis.Client {
	return redis.NewClient(&redis.Options{Addr: addr, Password: password, DB: dbNum})
}

func ParseRFC3339(v string) *time.Time {
	t, err := time.Parse(time.RFC3339, v)
	if err != nil {
		return nil
	}
	return &t
}
