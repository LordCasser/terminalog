import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, fireEvent, cleanup, act } from '@testing-library/react';
import { CommandPrompt } from './CommandPrompt';

// --- Mocks ---
const mockPush = vi.fn();
vi.mock('next/navigation', () => ({
  useRouter: () => ({ push: mockPush }),
  usePathname: () => '/',
}));

vi.mock('@/lib/hooks/useTerminalConfig', () => ({
  useTerminalConfig: () => ({ owner: 'testowner', currentDir: '' }),
}));

vi.mock('@/lib/api/articles', () => ({
  getArticles: vi.fn(),
}));

vi.mock('@/lib/api/search', () => ({
  searchArticles: vi.fn(),
}));

// Mock WebSocket
class MockWebSocket {
  static instances: MockWebSocket[] = [];
  readyState = 0; // CONNECTING
  onopen: (() => void) | null = null;
  onmessage: ((e: { data: string }) => void) | null = null;
  onclose: (() => void) | null = null;
  onerror: (() => void) | null = null;
  addEventListener = vi.fn();
  removeEventListener = vi.fn();
  send = vi.fn();
  close = vi.fn();
  constructor() {
    MockWebSocket.instances.push(this);
    // Simulate connection opening
    setTimeout(() => {
      this.readyState = 1; // OPEN
      this.onopen?.();
    }, 0);
  }
}

describe('CommandPrompt', () => {
  beforeEach(() => {
    global.WebSocket = MockWebSocket as unknown as typeof WebSocket;
    mockPush.mockClear();
  });

  afterEach(() => {
    cleanup();
    vi.restoreAllMocks();
  });

  describe('global keydown - trigger keys', () => {
    it('focuses input when "/" is pressed', () => {
      render(<CommandPrompt />);
      const input = screen.getByLabelText('Command input') as HTMLInputElement;
      expect(input).not.toBe(document.activeElement);

      act(() => {
        window.dispatchEvent(new KeyboardEvent('keydown', { key: '/' }));
      });

      expect(input).toBe(document.activeElement);
    });

    it('focuses input when ":" is pressed', () => {
      render(<CommandPrompt />);
      const input = screen.getByLabelText('Command input') as HTMLInputElement;

      act(() => {
        window.dispatchEvent(new KeyboardEvent('keydown', { key: ':' }));
      });

      expect(input).toBe(document.activeElement);
    });

    it('focuses input when Tab is pressed', () => {
      render(<CommandPrompt />);
      const input = screen.getByLabelText('Command input') as HTMLInputElement;

      act(() => {
        window.dispatchEvent(new KeyboardEvent('keydown', { key: 'Tab' }));
      });

      expect(input).toBe(document.activeElement);
    });
  });

  describe('global keydown - non-trigger keys do NOT focus', () => {
    it('does not focus input when letter "j" is pressed', () => {
      render(<CommandPrompt />);
      const input = screen.getByLabelText('Command input') as HTMLInputElement;

      act(() => {
        window.dispatchEvent(new KeyboardEvent('keydown', { key: 'j' }));
      });

      expect(input).not.toBe(document.activeElement);
    });

    it('does not focus input when letter "a" is pressed', () => {
      render(<CommandPrompt />);
      const input = screen.getByLabelText('Command input') as HTMLInputElement;

      act(() => {
        window.dispatchEvent(new KeyboardEvent('keydown', { key: 'a' }));
      });

      expect(input).not.toBe(document.activeElement);
    });

    it('does not focus input when modifier+key is pressed', () => {
      render(<CommandPrompt />);
      const input = screen.getByLabelText('Command input') as HTMLInputElement;

      act(() => {
        window.dispatchEvent(new KeyboardEvent('keydown', { key: '/', metaKey: true }));
      });

      expect(input).not.toBe(document.activeElement);
    });
  });

  describe('focused input - Escape and Ctrl+C', () => {
    it('blurs input on Escape', () => {
      render(<CommandPrompt />);
      const input = screen.getByLabelText('Command input') as HTMLInputElement;
      input.focus();
      expect(input).toBe(document.activeElement);

      fireEvent.keyDown(input, { key: 'Escape' });

      expect(input).not.toBe(document.activeElement);
    });

    it('clears input on Ctrl+C', () => {
      render(<CommandPrompt />);
      const input = screen.getByLabelText('Command input') as HTMLInputElement;
      input.focus();
      fireEvent.change(input, { target: { value: 'some text' } });
      expect(input.value).toBe('some text');

      fireEvent.keyDown(input, { key: 'c', ctrlKey: true });

      expect(input.value).toBe('');
    });
  });

  describe('commands: ls, pwd, clear', () => {
    it('ls navigates to root directory listing', () => {
      render(<CommandPrompt />);
      const input = screen.getByLabelText('Command input') as HTMLInputElement;
      input.focus();
      fireEvent.change(input, { target: { value: 'ls' } });
      fireEvent.submit(input.closest('form')!);

      expect(mockPush).toHaveBeenCalledWith('/');
    });

    it('clear empties input', () => {
      render(<CommandPrompt />);
      const input = screen.getByLabelText('Command input') as HTMLInputElement;
      input.focus();
      fireEvent.change(input, { target: { value: 'clear' } });
      // After submit, input should be cleared (handleSubmit clears input)
      fireEvent.submit(input.closest('form')!);

      expect(input.value).toBe('');
    });

    it('pwd shows transient info hint with current path', () => {
      vi.useFakeTimers();
      render(<CommandPrompt />);
      const input = screen.getByLabelText('Command input') as HTMLInputElement;
      input.focus();
      fireEvent.change(input, { target: { value: 'pwd' } });
      fireEvent.submit(input.closest('form')!);

      // Before pwd: only the footer path prefix shows "~/testowner"
      // After pwd: the info hint also shows "~/testowner" (2 occurrences)
      act(() => {
        vi.advanceTimersByTime(0);
      });
      const hints = screen.getAllByText('~/testowner');
      expect(hints.length).toBe(2);

      // After 2 seconds, the info hint disappears (back to 1 occurrence)
      act(() => {
        vi.advanceTimersByTime(2000);
      });
      const remaining = screen.queryAllByText('~/testowner');
      expect(remaining.length).toBe(1);

      vi.useRealTimers();
    });
  });
});
