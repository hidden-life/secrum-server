package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/hidden-life/secrum-server/internal/app/groups"
	"github.com/hidden-life/secrum-server/internal/app/messages"
)

func RegisterGroupsRoutes(r chi.Router, svc *groups.Service, msgSvc *messages.Service) {
	r.Route("/groups", func(r chi.Router) {
		r.Post("/", createGroupHandler(svc))
		r.Get("/", listGroupsHandler(svc))
		r.Get("/{group_id}/members", groupMembersList(svc))
		r.Post("/{group_id}/members", addGroupMember(svc))
		r.Delete("/{group_id}/members/{user_id}", removeGroupMember(svc))
		r.Post("/{group_id}/messages", sendGroupMessages(svc, msgSvc))
	})
}

func sendGroupMessages(svc *groups.Service, msgSvc *messages.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := UserIDFromContext(r.Context())
		deviceID := DeviceIDFromContext(r.Context())
		if userID == "" || deviceID == "" {
			asError(w, http.StatusUnauthorized, "unauthorized (missing user or device)")
			return
		}

		gid := chi.URLParam(r, "group_id")
		if gid == "" {
			asError(w, http.StatusNotFound, "group ID not found")
			return
		}

		var req messages.SendGroupMessageRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			asError(w, http.StatusBadRequest, "invalid input request")
			return
		}

		// get active members
		members, err := svc.GetActiveMemberIDs(r.Context(), userID, gid)
		if err != nil {
			asError(w, http.StatusBadRequest, err.Error())
			return
		}

		res, err := msgSvc.SendGroupMessage(r.Context(), userID, deviceID, gid, &req, members)
		if err != nil {
			asError(w, http.StatusBadRequest, err.Error())
			return
		}

		asJson(w, http.StatusOK, res)
	}
}

func removeGroupMember(svc *groups.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		actorUserID := UserIDFromContext(r.Context())
		if actorUserID == "" {
			asError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		gid := chi.URLParam(r, "group_id")
		if gid == "" {
			asError(w, http.StatusNotFound, "group not found")
			return
		}
		userID := chi.URLParam(r, "user_id")
		if userID == "" {
			asError(w, http.StatusNotFound, "user not found")
			return
		}

		if err := svc.RemoveMember(r.Context(), actorUserID, gid, userID); err != nil {
			asError(w, http.StatusBadRequest, err.Error())
			return
		}

		asJson(w, http.StatusNoContent, map[string]string{"status": "ok"})
	}
}

func addGroupMember(svc *groups.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		actorUserID := UserIDFromContext(r.Context())
		if actorUserID == "" {
			asError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		gid := chi.URLParam(r, "group_id")
		if gid == "" {
			asError(w, http.StatusNotFound, "group not found")
			return
		}

		var req groups.AddMemberRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			asError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if req.UserID == "" {
			asError(w, http.StatusBadRequest, "user_id not required field")
			return
		}

		if err := svc.AddMember(r.Context(), actorUserID, gid, req.UserID); err != nil {
			asError(w, http.StatusBadRequest, err.Error())
			return
		}

		asJson(w, http.StatusCreated, map[string]string{"status": "ok"})
	}
}

func groupMembersList(svc *groups.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := UserIDFromContext(r.Context())
		if userID == "" {
			asError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		gid := chi.URLParam(r, "group_id")
		if gid == "" {
			asError(w, http.StatusNotFound, "group not found")
			return
		}

		list, err := svc.ListMembers(r.Context(), userID, gid)
		if err != nil {
			asError(w, http.StatusBadRequest, err.Error())
			return
		}

		asJson(w, http.StatusOK, list)
	}
}

func listGroupsHandler(svc *groups.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := UserIDFromContext(r.Context())
		if userID == "" {
			asError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		list, err := svc.ListUserGroups(r.Context(), userID)
		if err != nil {
			asError(w, http.StatusBadRequest, err.Error())
			return
		}

		asJson(w, http.StatusOK, list)
	}
}

func createGroupHandler(svc *groups.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := UserIDFromContext(r.Context())
		if userID == "" {
			asError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		var req groups.CreateGroupRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			asError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		resp, err := svc.CreateGroup(r.Context(), userID, req.Name, req.AvatarURL, req.Members)
		if err != nil {
			asError(w, http.StatusBadRequest, err.Error())
			return
		}

		asJson(w, http.StatusCreated, resp)
	}
}
