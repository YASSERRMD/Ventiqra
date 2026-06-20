package server

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestCreateAndListEmployees(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	if rec := doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Team Co"}, token); rec.Code != http.StatusCreated {
		t.Fatalf("create company: %d body=%s", rec.Code, rec.Body.String())
	}

	// Empty list before any hires.
	rec := doJSON(t, srv, "GET", "/api/v1/companies/me/employees", nil, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("list empty status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var empty []employeeResponse
	if err := json.NewDecoder(rec.Body).Decode(&empty); err != nil {
		t.Fatalf("decode empty list: %v", err)
	}
	if empty == nil || len(empty) != 0 {
		t.Errorf("expected JSON [] not nil, got %v", empty)
	}

	// Hire with default salary (role default applied).
	rec = doJSON(t, srv, "POST", "/api/v1/companies/me/employees", map[string]any{"name": "Ada Lovelace", "role": "engineer"}, token)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create employee status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var e1 employeeResponse
	if err := json.NewDecoder(rec.Body).Decode(&e1); err != nil {
		t.Fatalf("decode employee: %v", err)
	}
	if e1.Name != "Ada Lovelace" || e1.Role != "engineer" {
		t.Errorf("employee = %+v", e1)
	}
	if e1.SalaryCents == 0 {
		t.Errorf("salary should default to non-zero for engineer")
	}
	if e1.Skill != 50 || e1.Morale != 70 {
		t.Errorf("skill/morale = %d/%d, want 50/70", e1.Skill, e1.Morale)
	}

	// Hire with explicit salary/skill/morale.
	rec = doJSON(t, srv, "POST", "/api/v1/companies/me/employees",
		map[string]any{"name": "Grace Hopper", "role": "designer", "salary_cents": 1100000, "skill": 88, "morale": 95}, token)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create second employee: %d body=%s", rec.Code, rec.Body.String())
	}
	var e2 employeeResponse
	_ = json.NewDecoder(rec.Body).Decode(&e2)
	if e2.SalaryCents != 1100000 || e2.Skill != 88 || e2.Morale != 95 {
		t.Errorf("explicit employee = %+v", e2)
	}

	// List returns both in hire order.
	rec = doJSON(t, srv, "GET", "/api/v1/companies/me/employees", nil, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("list status = %d", rec.Code)
	}
	var list []employeeResponse
	if err := json.NewDecoder(rec.Body).Decode(&list); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("list len = %d, want 2", len(list))
	}
	if list[0].ID != e1.ID || list[1].ID != e2.ID {
		t.Errorf("order mismatch")
	}
}

func TestCreateEmployeeValidation(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co"}, token)

	cases := []struct {
		name string
		body map[string]any
	}{
		{"missing name", map[string]any{"role": "engineer"}},
		{"empty name", map[string]any{"name": "", "role": "engineer"}},
		{"missing role", map[string]any{"name": "Someone"}},
		{"invalid role", map[string]any{"name": "Someone", "role": "ceo"}},
		{"negative salary", map[string]any{"name": "Someone", "role": "engineer", "salary_cents": -1}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			rec := doJSON(t, srv, "POST", "/api/v1/companies/me/employees", c.body, token)
			if rec.Code != http.StatusBadRequest {
				t.Errorf("%s: status = %d, want 400", c.name, rec.Code)
			}
		})
	}
}

func TestCreateEmployeeRequiresCompany(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	rec := doJSON(t, srv, "POST", "/api/v1/companies/me/employees", map[string]any{"name": "X", "role": "engineer"}, token)
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}
}

func TestCreateEmployeeRequiresAuth(t *testing.T) {
	srv := newAuthTestServer(t)
	rec := doJSON(t, srv, "POST", "/api/v1/companies/me/employees", map[string]any{"name": "X", "role": "engineer"}, "")
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}

func TestUpdateEmployeeSalary(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co"}, token)
	rec := doJSON(t, srv, "POST", "/api/v1/companies/me/employees", map[string]any{"name": "Raise", "role": "engineer"}, token)
	var e employeeResponse
	_ = json.NewDecoder(rec.Body).Decode(&e)

	rec = doJSON(t, srv, "PATCH", "/api/v1/employees/"+e.ID+"/salary", map[string]any{"salary_cents": 1500000}, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("update salary status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var updated employeeResponse
	if err := json.NewDecoder(rec.Body).Decode(&updated); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if updated.SalaryCents != 1500000 {
		t.Errorf("salary_cents = %d, want 1500000", updated.SalaryCents)
	}
}

func TestUpdateEmployeeMorale(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co"}, token)
	rec := doJSON(t, srv, "POST", "/api/v1/companies/me/employees", map[string]any{"name": "Moody", "role": "support"}, token)
	var e employeeResponse
	_ = json.NewDecoder(rec.Body).Decode(&e)

	rec = doJSON(t, srv, "PATCH", "/api/v1/employees/"+e.ID+"/morale", map[string]any{"morale": 99}, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("update morale status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var updated employeeResponse
	_ = json.NewDecoder(rec.Body).Decode(&updated)
	if updated.Morale != 99 {
		t.Errorf("morale = %d, want 99", updated.Morale)
	}
}

func TestDeleteEmployee(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co"}, token)
	rec := doJSON(t, srv, "POST", "/api/v1/companies/me/employees", map[string]any{"name": "Let Go", "role": "engineer"}, token)
	var e employeeResponse
	_ = json.NewDecoder(rec.Body).Decode(&e)

	rec = doJSON(t, srv, "DELETE", "/api/v1/employees/"+e.ID, nil, token)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("delete status = %d, want 204, body = %s", rec.Code, rec.Body.String())
	}
	// Second delete is 404.
	rec = doJSON(t, srv, "DELETE", "/api/v1/employees/"+e.ID, nil, token)
	if rec.Code != http.StatusNotFound {
		t.Errorf("second delete status = %d, want 404", rec.Code)
	}
}

func TestEmployeeForbiddenForOtherOwner(t *testing.T) {
	srv := newAuthTestServer(t)

	tokenA := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "A Co"}, tokenA)
	rec := doJSON(t, srv, "POST", "/api/v1/companies/me/employees", map[string]any{"name": "A Person", "role": "engineer"}, tokenA)
	var e employeeResponse
	_ = json.NewDecoder(rec.Body).Decode(&e)

	recB := doJSON(t, srv, "POST", "/api/v1/auth/register",
		map[string]string{"email": "other@example.com", "password": "password123", "name": "Other"}, "")
	if recB.Code != http.StatusOK {
		t.Fatalf("register B: %d", recB.Code)
	}
	var auth authResponse
	_ = json.NewDecoder(recB.Body).Decode(&auth)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "B Co"}, auth.Token)

	rec = doJSON(t, srv, "PATCH", "/api/v1/employees/"+e.ID+"/salary", map[string]any{"salary_cents": 500000}, auth.Token)
	if rec.Code != http.StatusForbidden {
		t.Errorf("cross-owner salary status = %d, want 403", rec.Code)
	}
	rec = doJSON(t, srv, "DELETE", "/api/v1/employees/"+e.ID, nil, auth.Token)
	if rec.Code != http.StatusForbidden {
		t.Errorf("cross-owner delete status = %d, want 403", rec.Code)
	}
}

func TestUpdateEmployeeMissingIs404(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co"}, token)

	missing := "00000000-0000-0000-0000-000000000000"
	rec := doJSON(t, srv, "PATCH", "/api/v1/employees/"+missing+"/salary", map[string]any{"salary_cents": 100000}, token)
	if rec.Code != http.StatusNotFound {
		t.Errorf("salary missing status = %d, want 404", rec.Code)
	}
	rec = doJSON(t, srv, "PATCH", "/api/v1/employees/"+missing+"/morale", map[string]any{"morale": 50}, token)
	if rec.Code != http.StatusNotFound {
		t.Errorf("morale missing status = %d, want 404", rec.Code)
	}
	rec = doJSON(t, srv, "DELETE", "/api/v1/employees/"+missing, nil, token)
	if rec.Code != http.StatusNotFound {
		t.Errorf("delete missing status = %d, want 404", rec.Code)
	}
}
