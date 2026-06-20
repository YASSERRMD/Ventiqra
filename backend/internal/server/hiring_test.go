package server

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestListCandidates(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Hiring Co"}, token)

	rec := doJSON(t, srv, "GET", "/api/v1/companies/me/candidates", nil, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("list candidates status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var pool candidatePoolResponse
	if err := json.NewDecoder(rec.Body).Decode(&pool); err != nil {
		t.Fatalf("decode pool: %v", err)
	}
	if len(pool.Candidates) != 6 {
		t.Fatalf("pool size = %d, want 6", len(pool.Candidates))
	}
	for i, c := range pool.Candidates {
		if c.Index != i {
			t.Errorf("candidate %d index = %d", i, c.Index)
		}
		if c.HiringFeeCents <= 0 || c.SalaryExpectationCents <= 0 {
			t.Errorf("candidate %d has non-positive economics: %+v", i, c)
		}
	}
}

func TestCandidatesAreDeterministicPerRound(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co"}, token)

	first := doJSON(t, srv, "GET", "/api/v1/companies/me/candidates", nil, token)
	second := doJSON(t, srv, "GET", "/api/v1/companies/me/candidates", nil, token)
	if first.Body.String() != second.Body.String() {
		t.Errorf("candidate pool is not deterministic across calls")
	}
}

func TestHireCandidateDeterministicDecision(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co"}, token)

	a := doJSON(t, srv, "POST", "/api/v1/companies/me/candidates/0/hire", nil, token)
	b := doJSON(t, srv, "POST", "/api/v1/companies/me/candidates/0/hire", nil, token)
	if a.Code != http.StatusOK && a.Code != http.StatusCreated {
		t.Fatalf("first hire status = %d, body = %s", a.Code, a.Body.String())
	}
	var ra, rb hireResultResponse
	_ = json.NewDecoder(a.Body).Decode(&ra)
	_ = json.NewDecoder(b.Body).Decode(&rb)
	if ra.Accepted != rb.Accepted {
		t.Errorf("offer decision for index 0 is not deterministic: %v vs %v", ra.Accepted, rb.Accepted)
	}
}

func TestHireCandidateEffects(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co"}, token)

	// Baseline cash.
	coBefore := doJSON(t, srv, "GET", "/api/v1/companies/me", nil, token)
	var before companyResponse
	_ = json.NewDecoder(coBefore.Body).Decode(&before)

	// Hire all six candidates, tracking accepted fees.
	poolRec := doJSON(t, srv, "GET", "/api/v1/companies/me/candidates", nil, token)
	var pool candidatePoolResponse
	_ = json.NewDecoder(poolRec.Body).Decode(&pool)

	totalFees := int64(0)
	acceptedCount := 0
	for i := 0; i < len(pool.Candidates); i++ {
		rec := doJSON(t, srv, "POST", "/api/v1/companies/me/candidates/"+itoa(i)+"/hire", nil, token)
		if rec.Code != http.StatusOK && rec.Code != http.StatusCreated {
			t.Fatalf("hire %d status = %d, body = %s", i, rec.Code, rec.Body.String())
		}
		var res hireResultResponse
		if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
			t.Fatalf("decode hire %d: %v", i, err)
		}
		if res.Accepted {
			acceptedCount++
			totalFees += pool.Candidates[i].HiringFeeCents
			if res.Employee == nil {
				t.Errorf("accepted hire %d returned no employee", i)
			}
		}
	}

	// Team size grew by the accepted count.
	list := doJSON(t, srv, "GET", "/api/v1/companies/me/employees", nil, token)
	var team []employeeResponse
	_ = json.NewDecoder(list.Body).Decode(&team)
	if len(team) != acceptedCount {
		t.Errorf("team size = %d, want %d", len(team), acceptedCount)
	}

	// Cash decreased by exactly the accepted fees.
	coAfter := doJSON(t, srv, "GET", "/api/v1/companies/me", nil, token)
	var after companyResponse
	_ = json.NewDecoder(coAfter.Body).Decode(&after)
	if after.CashCents != before.CashCents-totalFees {
		t.Errorf("cash = %d, want %d (before %d - fees %d)",
			after.CashCents, before.CashCents-totalFees, before.CashCents, totalFees)
	}
}

func TestHireInvalidIndex(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	doJSON(t, srv, "POST", "/api/v1/companies", map[string]any{"name": "Co"}, token)

	cases := []string{"-1", "6", "abc"}
	for _, idx := range cases {
		rec := doJSON(t, srv, "POST", "/api/v1/companies/me/candidates/"+idx+"/hire", nil, token)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("index %q: status = %d, want 400", idx, rec.Code)
		}
	}
}

func TestCandidatesRequireCompany(t *testing.T) {
	srv := newAuthTestServer(t)
	token := registerAndLogin(t, srv)
	rec := doJSON(t, srv, "GET", "/api/v1/companies/me/candidates", nil, token)
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}
}

func TestCandidatesRequireAuth(t *testing.T) {
	srv := newAuthTestServer(t)
	rec := doJSON(t, srv, "GET", "/api/v1/companies/me/candidates", nil, "")
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}

// itoa is a tiny strconv-free helper to keep imports minimal in test files.
func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	var buf [20]byte
	pos := len(buf)
	neg := i < 0
	if neg {
		i = -i
	}
	for i > 0 {
		pos--
		buf[pos] = byte('0' + i%10)
		i /= 10
	}
	if neg {
		pos--
		buf[pos] = '-'
	}
	return string(buf[pos:])
}
