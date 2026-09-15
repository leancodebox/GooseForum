import { Activity } from "react";
import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";
import type {
  AnyPagePayload,
  GooseSiteApi,
  HomeProps,
  SearchPageProps,
  TopicPayload,
} from "@gooseforum/client";
import { GooseI18nProvider } from "@gooseforum/runtime/i18n";
import {
  GooseRuntimeProvider,
  type GooseRuntime,
} from "@gooseforum/runtime";
import { SearchPageView } from "../src/site/pages/search-page";
import { useTopicList } from "../src/site/topics/topic-list";
import { Calendar } from "@gooseforum/ui/components/calendar";

afterEach(cleanup);

describe("topic list keep-alive state", () => {
  it("uses unique search input IDs across cached page instances", () => {
    const page: SearchPageProps = {
      query: "", topics: [], total: 0, totalPages: 0,
      pagination: { page: 1, nextPage: 2, hasNext: false, nextUrl: "" },
    };
    render(<GooseI18nProvider locale="zh"><GooseRuntimeProvider runtime={createRuntime()}>
      <SearchPageView page={page} /><SearchPageView page={page} />
    </GooseRuntimeProvider></GooseI18nProvider>);
    const inputs = screen.getAllByRole("textbox") as HTMLInputElement[];
    expect(inputs[0].id).not.toBe(inputs[1].id);
    for (const input of inputs) {
      expect(input.labels?.[0].control).toBe(input);
    }
  });

  it("keeps the calendar root DOM stable on parent updates", () => {
    const renderCalendar = (className: string) => <Calendar mode="single" month={new Date(2026, 8)} className={className} />;
    const view = render(renderCalendar("first"));
    const root = document.querySelector('[data-slot="calendar"]');
    view.rerender(renderCalendar("second"));
    expect(document.querySelector('[data-slot="calendar"]')).toBe(root);
  });
  it("does not recreate the list observer for unrelated runtime updates", () => {
    const observe = vi.fn();
    const disconnect = vi.fn();
    vi.stubGlobal("IntersectionObserver", class {
      observe = observe;
      disconnect = disconnect;
    });
    try {
      const page = topicPage([topic(1, "First")], {
        page: 1, nextPage: 2, hasNext: true, nextUrl: "/?page=2",
      });
      const runtime = createRuntime();
      const view = (value: GooseRuntime) => (
        <GooseRuntimeProvider runtime={value}><TopicListState page={page} /></GooseRuntimeProvider>
      );
      const rendered = render(view(runtime));
      const before = observe.mock.calls.length;
      expect(before).toBeGreaterThan(0);
      rendered.rerender(view({ ...runtime, isNavigating: true }));
      expect(observe).toHaveBeenCalledTimes(before);
      rendered.unmount();
    } finally {
      vi.unstubAllGlobals();
    }
  });
  it("keeps appended topics when an Activity is hidden and shown again", async () => {
    const firstPage = topicPage([topic(1, "First")], {
      page: 1,
      nextPage: 2,
      hasNext: true,
      nextUrl: "/?page=2",
    });
    const secondPage = topicPage([topic(2, "Second")], {
      page: 2,
      nextPage: 3,
      hasNext: false,
      nextUrl: "",
    });
    const runtime = createRuntime(async () => payload(secondPage, "/?page=2"));
    const view = (mode: "visible" | "hidden", page: HomeProps) => (
      <GooseRuntimeProvider runtime={runtime}>
        <Activity mode={mode}>
          <TopicListState page={page} />
        </Activity>
      </GooseRuntimeProvider>
    );
    const rendered = render(view("visible", firstPage));

    fireEvent.click(screen.getByRole("button", { name: "load more" }));
    await waitFor(() => expect(screen.getByTestId("topics").textContent).toBe("1,2"));

    rendered.rerender(view("hidden", { ...firstPage }));
    rendered.rerender(view("visible", { ...firstPage }));

    await waitFor(() => expect(screen.getByTestId("topics").textContent).toBe("1,2"));
  });

  it("keeps an unfinished search query when an Activity is shown again", async () => {
    const page: SearchPageProps = {
      query: "initial",
      topics: [],
      total: 0,
      totalPages: 0,
      pagination: { page: 1, nextPage: 2, hasNext: false, nextUrl: "" },
    };
    const runtime = createRuntime();
    const view = (mode: "visible" | "hidden") => (
      <GooseI18nProvider locale="zh">
        <GooseRuntimeProvider runtime={runtime}>
          <Activity mode={mode}>
            <SearchPageView page={{ ...page }} />
          </Activity>
        </GooseRuntimeProvider>
      </GooseI18nProvider>
    );
    const rendered = render(view("visible"));
    const input = screen.getByRole("textbox");

    await userEvent.clear(input);
    await userEvent.type(input, "unfinished");
    rendered.rerender(view("hidden"));
    rendered.rerender(view("visible"));

    expect((screen.getByRole("textbox") as HTMLInputElement).value).toBe("unfinished");
  });
});

function TopicListState({ page }: { page: HomeProps }) {
  const list = useTopicList(page, "/");
  return (
    <div>
      <div ref={list.sentinel} />
      <output data-testid="topics">{list.topics.map((item) => item.id).join(",")}</output>
      <button type="button" onClick={() => void list.loadMore()}>
        load more
      </button>
    </div>
  );
}

function createRuntime(
  fetchPage: NonNullable<GooseRuntime["fetchPage"]> = vi.fn(),
): GooseRuntime {
  return {
    api: {} as GooseSiteApi,
    currentUrl: "/",
    isNavigating: false,
    theme: "gf-light",
    locale: "zh",
    navigate: vi.fn(async () => {}),
    fetchPage,
    queueFlash: vi.fn(),
    redirect: vi.fn(),
    refresh: vi.fn(async () => {}),
    setLocale: vi.fn(),
    toggleTheme: vi.fn(),
  };
}

function payload(page: HomeProps, url: string): AnyPagePayload {
  return {
    component: "home.index",
    props: page,
    layout: {} as AnyPagePayload["layout"],
    meta: { title: "Test" },
    url,
    version: "1",
  };
}

function topicPage(
  topics: TopicPayload[],
  pagination: HomeProps["pagination"],
): HomeProps {
  return {
    sort: "latest",
    tabs: [],
    topics,
    pagination,
    announcement: { enabled: false, html: "" },
  };
}

function topic(id: number, title: string): TopicPayload {
  return {
    id,
    title,
    description: "",
    url: `/t/${id}`,
    author: { id, username: title, avatarUrl: "" },
    participants: [],
    categories: [],
    replyCount: 0,
    viewCount: 0,
    pinWeight: 0,
    processStatus: 0,
    activityText: "",
    lastUpdateTime: "2026-09-14 00:00:00",
  };
}
