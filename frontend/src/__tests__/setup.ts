import "@testing-library/jest-dom/vitest";

// Stub the WebSocket constructor so useRealtime doesn't blow up in tests.
class MockWebSocket {
  static instances: MockWebSocket[] = [];
  readyState = 0;
  onopen: (() => void) | null = null;
  onmessage: ((ev: { data: string }) => void) | null = null;
  onclose: (() => void) | null = null;
  onerror: (() => void) | null = null;
  constructor(public url: string) {
    MockWebSocket.instances.push(this);
  }
  close() {
    this.readyState = 3;
    this.onclose?.();
  }
  send() {}
}
// @ts-expect-error override global
global.WebSocket = MockWebSocket;
