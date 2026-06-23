"use client";

import { Component, ReactNode } from "react";

type Props = { children: ReactNode };
type State = { hasError: boolean; message: string };

/**
 * ErrorBoundary catches render-time errors in its children and shows a
 * user-friendly fallback instead of a blank page.
 */
export class ErrorBoundary extends Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = { hasError: false, message: "" };
  }

  static getDerivedStateFromError(error: Error): State {
    return { hasError: true, message: error.message };
  }

  render() {
    if (this.state.hasError) {
      return (
        <div className="flex min-h-[200px] flex-col items-center justify-center rounded-xl border border-rose-500/40 bg-rose-500/10 p-8 text-center">
          <p className="text-sm font-semibold text-rose-400">Something went wrong</p>
          <p className="mt-1 max-w-md text-xs text-muted">
            {this.state.message || "An unexpected error occurred."}
          </p>
          <button
            onClick={() => this.setState({ hasError: false, message: "" })}
            className="mt-4 rounded-md bg-brand px-4 py-2 text-xs font-semibold text-surface-muted"
          >
            Try again
          </button>
        </div>
      );
    }
    return this.props.children;
  }
}
