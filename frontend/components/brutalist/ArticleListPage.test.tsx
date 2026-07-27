import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, waitFor, fireEvent, cleanup } from "@testing-library/react";
import { ArticleListPage } from "./ArticleListPage";
import { getArticles } from "@/lib/api/articles";
import { ApiError } from "@/lib/api/client";

vi.mock("@/lib/api/articles", () => ({
  getArticles: vi.fn(),
}));

const mockedGetArticles = vi.mocked(getArticles);

beforeEach(() => {
  vi.clearAllMocks();
});

afterEach(() => {
  cleanup();
});

describe("ArticleListPage", () => {
  it("renders skeleton while loading", () => {
    // Never-resolving promise keeps loading=true
    mockedGetArticles.mockReturnValue(new Promise(() => {}));
    render(<ArticleListPage currentDir="" />);
    const skeleton = document.querySelector('[class*="min-w-[480px]"]');
    expect(skeleton).not.toBeNull();
    expect(mockedGetArticles).toHaveBeenCalledTimes(1);
  });

  it("renders empty state when directory has no entries", async () => {
    mockedGetArticles.mockResolvedValue({ articles: [], total: 0, currentDir: "" });
    render(<ArticleListPage currentDir="" />);
    await waitFor(() => {
      expect(screen.getByText("no entries")).not.toBeNull();
    });
  });

  it("renders error state with retry button when fetch fails (non-404)", async () => {
    mockedGetArticles.mockRejectedValue(new Error("network"));
    render(<ArticleListPage currentDir="" />);
    await waitFor(() => {
      expect(screen.getByText("Directory Unavailable")).not.toBeNull();
      expect(screen.getByText("retry")).not.toBeNull();
    });
  });

  it("renders not-found state without retry on 404", async () => {
    mockedGetArticles.mockRejectedValue(new ApiError("Not found", 404));
    render(<ArticleListPage currentDir="" />);
    await waitFor(() => {
      expect(screen.getByText("Not Found")).not.toBeNull();
      expect(
        screen.getByText("The requested directory does not exist.")
      ).not.toBeNull();
    });
    // No retry button for 404
    expect(screen.queryByText("retry")).toBeNull();
  });

  it("retries fetch on retry click", async () => {
    mockedGetArticles.mockRejectedValueOnce(new Error("network"));
    render(<ArticleListPage currentDir="" />);
    await waitFor(() => {
      expect(screen.getByText("retry")).not.toBeNull();
    });
    // Second call succeeds with empty directory
    mockedGetArticles.mockResolvedValue({ articles: [], total: 0, currentDir: "" });
    fireEvent.click(screen.getByText("retry"));
    await waitFor(() => {
      expect(screen.getByText("no entries")).not.toBeNull();
    });
    expect(mockedGetArticles).toHaveBeenCalledTimes(2);
  });
});

