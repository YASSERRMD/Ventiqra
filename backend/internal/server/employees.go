package server

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/YASSERRMD/Ventiqra/backend/internal/middleware"
	"github.com/YASSERRMD/Ventiqra/backend/internal/repository"
)

type employeeResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Role        string    `json:"role"`
	SalaryCents int64     `json:"salary_cents"`
	Skill       int       `json:"skill"`
	Morale      int       `json:"morale"`
	HiredAt     time.Time `json:"hired_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type createEmployeeRequest struct {
	Name        string `json:"name"`
	Role        string `json:"role"`
	SalaryCents *int64 `json:"salary_cents"`
	Skill       *int   `json:"skill"`
	Morale      *int   `json:"morale"`
}

type updateEmployeeSalaryRequest struct {
	SalaryCents *int64 `json:"salary_cents"`
}

type updateEmployeeMoraleRequest struct {
	Morale *int `json:"morale"`
}

func toEmployeeResponse(e *repository.Employee) employeeResponse {
	return employeeResponse{
		ID: e.ID, Name: e.Name, Role: string(e.Role), SalaryCents: e.SalaryCents,
		Skill: e.Skill, Morale: e.Morale, HiredAt: e.HiredAt,
		CreatedAt: e.CreatedAt, UpdatedAt: e.UpdatedAt,
	}
}

// handleCreateEmployee hires a new employee for the owner's latest company.
func (s *Server) handleCreateEmployee(w http.ResponseWriter, r *http.Request) {
	if s.employees == nil || s.companies == nil {
		writeError(w, http.StatusServiceUnavailable, "employee service not configured")
		return
	}

	var req createEmployeeRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if len(req.Name) < 1 || len(req.Name) > 120 {
		writeError(w, http.StatusBadRequest, "name must be 1-120 characters")
		return
	}
	role := repository.EmployeeRole(strings.TrimSpace(req.Role))
	if !repository.ValidEmployeeRoles[role] {
		writeError(w, http.StatusBadRequest, "role must be one of engineer, designer, sales, marketing, support, operations")
		return
	}

	companyID, ok := s.ownerCompanyID(w, r)
	if !ok {
		return
	}

	var salary int64
	if req.SalaryCents != nil {
		salary = *req.SalaryCents
		if salary < 0 {
			writeError(w, http.StatusBadRequest, "salary_cents must be non-negative")
			return
		}
	}
	skill, morale := -1, -1
	if req.Skill != nil {
		skill = *req.Skill
	}
	if req.Morale != nil {
		morale = *req.Morale
	}

	emp, err := s.employees.CreateEmployee(r.Context(), companyID, req.Name, role, salary, skill, morale)
	if err != nil {
		if errors.Is(err, repository.ErrConflict) {
			writeError(w, http.StatusConflict, "employee conflict")
			return
		}
		s.log.Error("create employee failed", "error", err)
		writeError(w, http.StatusInternalServerError, "could not hire employee")
		return
	}
	writeJSON(w, http.StatusCreated, toEmployeeResponse(emp))
}

// handleListEmployees returns the team for the owner's latest company.
func (s *Server) handleListEmployees(w http.ResponseWriter, r *http.Request) {
	if s.employees == nil || s.companies == nil {
		writeError(w, http.StatusServiceUnavailable, "employee service not configured")
		return
	}

	companyID, ok := s.ownerCompanyID(w, r)
	if !ok {
		return
	}

	list, err := s.employees.ListEmployeesByCompany(r.Context(), companyID)
	if err != nil {
		s.log.Error("list employees failed", "error", err)
		writeError(w, http.StatusInternalServerError, "could not load employees")
		return
	}
	if list == nil {
		list = []*repository.Employee{}
	}
	resp := make([]employeeResponse, 0, len(list))
	for _, e := range list {
		resp = append(resp, toEmployeeResponse(e))
	}
	writeJSON(w, http.StatusOK, resp)
}

// handleUpdateEmployeeSalary adjusts an employee's monthly salary.
func (s *Server) handleUpdateEmployeeSalary(w http.ResponseWriter, r *http.Request) {
	if s.employees == nil || s.companies == nil {
		writeError(w, http.StatusServiceUnavailable, "employee service not configured")
		return
	}

	var req updateEmployeeSalaryRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.SalaryCents == nil || *req.SalaryCents < 0 {
		writeError(w, http.StatusBadRequest, "salary_cents must be a non-negative integer")
		return
	}

	if _, ok := s.loadOwnedEmployee(w, r); !ok {
		return
	}
	id := r.PathValue("id")
	if err := s.employees.UpdateSalary(r.Context(), id, *req.SalaryCents); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "employee not found")
			return
		}
		s.log.Error("update salary failed", "error", err)
		writeError(w, http.StatusInternalServerError, "could not update salary")
		return
	}
	updated, err := s.employees.GetEmployee(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{"id": id, "salary_cents": *req.SalaryCents})
		return
	}
	writeJSON(w, http.StatusOK, toEmployeeResponse(updated))
}

// handleUpdateEmployeeMorale adjusts an employee's morale.
func (s *Server) handleUpdateEmployeeMorale(w http.ResponseWriter, r *http.Request) {
	if s.employees == nil || s.companies == nil {
		writeError(w, http.StatusServiceUnavailable, "employee service not configured")
		return
	}

	var req updateEmployeeMoraleRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.Morale == nil {
		writeError(w, http.StatusBadRequest, "morale is required")
		return
	}

	if _, ok := s.loadOwnedEmployee(w, r); !ok {
		return
	}
	id := r.PathValue("id")
	if err := s.employees.UpdateMorale(r.Context(), id, *req.Morale); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "employee not found")
			return
		}
		s.log.Error("update morale failed", "error", err)
		writeError(w, http.StatusInternalServerError, "could not update morale")
		return
	}
	updated, err := s.employees.GetEmployee(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{"id": id, "morale": *req.Morale})
		return
	}
	writeJSON(w, http.StatusOK, toEmployeeResponse(updated))
}

// handleDeleteEmployee fires an employee.
func (s *Server) handleDeleteEmployee(w http.ResponseWriter, r *http.Request) {
	if s.employees == nil || s.companies == nil {
		writeError(w, http.StatusServiceUnavailable, "employee service not configured")
		return
	}
	if _, ok := s.loadOwnedEmployee(w, r); !ok {
		return
	}
	id := r.PathValue("id")
	if err := s.employees.DeleteEmployee(r.Context(), id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "employee not found")
			return
		}
		s.log.Error("delete employee failed", "error", err)
		writeError(w, http.StatusInternalServerError, "could not fire employee")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// loadOwnedEmployee loads an employee by path id and confirms it belongs to the
// authenticated owner's latest company.
func (s *Server) loadOwnedEmployee(w http.ResponseWriter, r *http.Request) (*repository.Employee, bool) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "employee id is required")
		return nil, false
	}
	ownerID := middleware.UserIDFrom(r.Context())
	company, err := s.companies.GetLatestCompanyForOwner(r.Context(), ownerID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "no company found")
			return nil, false
		}
		s.log.Error("load company failed", "error", err)
		writeError(w, http.StatusInternalServerError, "could not load company")
		return nil, false
	}
	emp, err := s.employees.GetEmployee(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "employee not found")
			return nil, false
		}
		s.log.Error("load employee failed", "error", err)
		writeError(w, http.StatusInternalServerError, "could not load employee")
		return nil, false
	}
	if emp.CompanyID != company.ID {
		writeError(w, http.StatusForbidden, "forbidden")
		return nil, false
	}
	return emp, true
}
