import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, fireEvent, cleanup, act } from '@testing-library/react';
import { HelpModal, SHOW_HELP_MODAL } from './HelpModal';

describe('HelpModal auto-close', () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
    cleanup();
  });

  it('closes after 10s with no interaction', () => {
    render(<HelpModal />);
    expect(screen.queryByText('Available Commands')).toBeNull();

    act(() => {
      window.dispatchEvent(new Event(SHOW_HELP_MODAL));
    });
    expect(screen.queryByText('Available Commands')).not.toBeNull();

    // 9s: still open
    act(() => {
      vi.advanceTimersByTime(9000);
    });
    expect(screen.queryByText('Available Commands')).not.toBeNull();

    // 10s: closes
    act(() => {
      vi.advanceTimersByTime(1000);
    });
    expect(screen.queryByText('Available Commands')).toBeNull();
  });

  it('stays open when the pointer moves over the modal (timer reset)', () => {
    render(<HelpModal />);
    act(() => {
      window.dispatchEvent(new Event(SHOW_HELP_MODAL));
    });
    expect(screen.queryByText('Available Commands')).not.toBeNull();

    // 9s pass
    act(() => {
      vi.advanceTimersByTime(9000);
    });
    expect(screen.queryByText('Available Commands')).not.toBeNull();

    // User moves the pointer over the modal content (any descendant triggers
    // the React onMouseMove bound on the modal panel). This resets the timer.
    const footer = screen.getByText(/Auto-closes after 10s idle/);
    act(() => {
      fireEvent.mouseMove(footer);
    });

    // Another 9s pass (18s total since open, but only 9s since last interaction)
    act(() => {
      vi.advanceTimersByTime(9000);
    });
    expect(screen.queryByText('Available Commands')).not.toBeNull();

    // 2 more seconds -> 11s since last interaction -> closes
    act(() => {
      vi.advanceTimersByTime(2000);
    });
    expect(screen.queryByText('Available Commands')).toBeNull();
  });

  it('manual close clears the timer and does not fire later', () => {
    render(<HelpModal />);
    act(() => {
      window.dispatchEvent(new Event(SHOW_HELP_MODAL));
    });

    // Close via × button
    const closeBtn = screen.getByLabelText('Close help modal');
    act(() => {
      fireEvent.click(closeBtn);
    });
    expect(screen.queryByText('Available Commands')).toBeNull();

    // Advance well past 10s; ensure no state update / error
    act(() => {
      vi.advanceTimersByTime(15000);
    });
    expect(screen.queryByText('Available Commands')).toBeNull();
  });
});
