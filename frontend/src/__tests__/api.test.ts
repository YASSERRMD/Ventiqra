import { describe, it, expect, vi, beforeEach } from "vitest";
import { api, ApiError } from "@/lib/api";

const fetchMock = vi.fn();
vi.stubGlobal("fetch", fetchMock);

const localStorageMock = {
  getItem: vi.fn(),
  setItem: vi.fn(),
  removeItem: vi.fn(),
  clear: vi.fn(),
};
vi.stubGlobal("localStorage", localStorageMock);

function jsonResponse(body: unknown, status = 200): Response {
  const text = JSON.stringify(body);
  return {
    ok: status >= 200 && status < 300,
    status,
    text: () => Promise.resolve(text),
    json: () => Promise.resolve(body),
    headers: new Headers(),
  } as Response;
}

describe("api client", () => {
  beforeEach(() => {
    fetchMock.mockReset();
    localStorageMock.getItem.mockReturnValue(null);
  });

  it("GET returns parsed JSON on success", async () => {
    fetchMock.mockResolvedValue(jsonResponse({ day: 5 }));
    const result = await api.get("/api/v1/test", { token: "tok" });
    expect(result).toEqual({ day: 5 });
    const [url, init] = fetchMock.mock.calls[0];
    expect(url).toContain("/api/v1/test");
    expect(init.method).toBe("GET");
  });

  it("POST sends a JSON body", async () => {
    fetchMock.mockResolvedValue(jsonResponse({ ok: true }, 201));
    await api.post("/api/v1/test", { name: "Co" }, { token: "tok" });
    const [, init] = fetchMock.mock.calls[0];
    expect(init.method).toBe("POST");
    expect(init.body).toBe(JSON.stringify({ name: "Co" }));
  });

  it("throws ApiError on non-2xx", async () => {
    fetchMock.mockResolvedValue(jsonResponse({ error: "bad" }, 400));
    try {
      await api.get("/api/v1/test", { token: "tok" });
      expect.fail("should have thrown");
    } catch (e) {
      expect(e).toBeInstanceOf(ApiError);
      expect((e as ApiError).status).toBe(400);
    }
  });

  it("DELETE sends DELETE method", async () => {
    fetchMock.mockResolvedValue(jsonResponse({}, 200));
    await api.del("/api/v1/test/1", { token: "tok" });
    const [, init] = fetchMock.mock.calls[0];
    expect(init.method).toBe("DELETE");
  });

  it("PATCH sends PATCH method", async () => {
    fetchMock.mockResolvedValue(jsonResponse({}, 200));
    await api.patch("/api/v1/test/1", { name: "new" }, { token: "tok" });
    const [, init] = fetchMock.mock.calls[0];
    expect(init.method).toBe("PATCH");
  });
});

describe("ApiError", () => {
  it("carries status and body", () => {
    const e = new ApiError(404, "not found", { detail: "x" });
    expect(e.status).toBe(404);
    expect(e.body).toEqual({ detail: "x" });
    expect(e.message).toBe("not found");
  });
});
