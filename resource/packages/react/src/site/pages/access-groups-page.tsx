import { useCallback, useEffect, useState } from "react";
import type {
  JoinableAccessGroup,
  ManagedAccessGroup,
} from "@gooseforum/client";
import { CheckCircle2, Clock3, Lock, UsersRound } from "lucide-react";
import { useTranslation } from "react-i18next";
import { Alert, AlertDescription } from "../../components/ui/alert";
import { Button } from "../../components/ui/button";
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "../../components/ui/empty";
import { Spinner } from "../../components/ui/spinner";
import { useGooseRuntime } from "../../runtime";
import { PageHeader } from "../layout/page-header";

export function AccessGroupsPageView() {
  const { t } = useTranslation("accessGroups");
  const { t: commonT } = useTranslation("contentCommon");
  const { api } = useGooseRuntime();
  const [groups, setGroups] = useState<JoinableAccessGroup[]>([]);
  const [managed, setManaged] = useState<ManagedAccessGroup[]>([]);
  const [loading, setLoading] = useState(true);
  const [applyingId, setApplyingId] = useState(0);
  const [reviewingId, setReviewingId] = useState(0);
  const [error, setError] = useState("");

  const load = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const [nextGroups, nextManaged] = await Promise.all([
        api.accessGroups.list(),
        api.accessGroups.managed(),
      ]);
      setGroups(nextGroups);
      setManaged(nextManaged);
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : t("loadFailed"));
    } finally {
      setLoading(false);
    }
  }, [api.accessGroups, t]);
  useEffect(() => {
    void load();
  }, [load]);

  async function apply(group: JoinableAccessGroup) {
    if (applyingId) return;
    setApplyingId(group.id);
    setError("");
    try {
      await api.accessGroups.apply(group.id);
      setGroups((items) =>
        items.map((item) =>
          item.id === group.id ? { ...item, status: 2 } : item,
        ),
      );
    } catch (reason) {
      setError(
        reason instanceof Error ? reason.message : t("applicationFailed"),
      );
    } finally {
      setApplyingId(0);
    }
  }
  async function review(groupId: number, memberId: number, approve: boolean) {
    if (reviewingId) return;
    setReviewingId(memberId);
    setError("");
    try {
      await api.accessGroups.review(groupId, memberId, approve);
      setManaged(await api.accessGroups.managed());
    } catch (reason) {
      setError(
        reason instanceof Error ? reason.message : t("memberSaveFailed"),
      );
    } finally {
      setReviewingId(0);
    }
  }

  return (
    <main className="min-w-0 pb-8">
      <PageHeader title={t("joinTitle")} description={t("joinDescription")} />
      {error ? (
        <Alert variant="destructive" className="mb-3">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      ) : null}
      <section className="overflow-hidden rounded-xl border bg-background">
        {loading ? (
          <div className="flex min-h-40 items-center justify-center gap-2 text-sm text-muted-foreground">
            <Spinner />
            {commonT("loading")}
          </div>
        ) : groups.length ? (
          <div className="divide-y">
            {groups.map((group) => (
              <article
                key={group.id}
                className="flex flex-col gap-4 px-5 py-4 sm:flex-row sm:items-center sm:justify-between"
              >
                <div className="min-w-0">
                  <h2 className="flex items-center gap-2 font-semibold">
                    <Lock />
                    {group.name}
                  </h2>
                  <p className="mt-1 text-sm text-muted-foreground">
                    {group.categories.length
                      ? t("unlocksCategories", {
                          categories: group.categories.join("、"),
                        })
                      : t("noCategoryGrants")}
                  </p>
                </div>
                {group.status === 1 ? (
                  <span className="inline-flex items-center gap-1.5 text-sm font-medium text-emerald-600">
                    <CheckCircle2 />
                    {t("joined")}
                  </span>
                ) : group.status === 2 ? (
                  <span className="inline-flex items-center gap-1.5 text-sm font-medium text-amber-600">
                    <Clock3 />
                    {t("pending")}
                  </span>
                ) : (
                  <Button
                    disabled={Boolean(applyingId)}
                    onClick={() => void apply(group)}
                  >
                    {applyingId === group.id ? (
                      <Spinner data-icon="inline-start" />
                    ) : null}
                    {applyingId === group.id ? commonT("saving") : t("apply")}
                  </Button>
                )}
              </article>
            ))}
          </div>
        ) : (
          <Empty className="min-h-48 border-0">
            <EmptyHeader>
              <EmptyMedia variant="icon">
                <UsersRound />
              </EmptyMedia>
              <EmptyTitle>{t("noJoinableGroups")}</EmptyTitle>
              <EmptyDescription>{t("noJoinableGroupsHint")}</EmptyDescription>
            </EmptyHeader>
          </Empty>
        )}
      </section>
      {managed.length ? (
        <section className="mt-4 overflow-hidden rounded-xl border bg-background">
          <header className="border-b px-5 py-4">
            <h2 className="font-semibold">{t("applicationsToReview")}</h2>
            <p className="mt-1 text-sm text-muted-foreground">
              {t("managerReviewHint")}
            </p>
          </header>
          <div className="divide-y">
            {managed.map((group) => (
              <article key={group.id} className="px-5 py-4">
                <h3 className="font-medium">{group.name}</h3>
                {group.applications.length ? (
                  group.applications.map((application) => (
                    <div
                      key={application.id}
                      className="mt-3 flex items-center justify-between gap-3 rounded-lg border p-3"
                    >
                      <span className="font-medium">
                        {application.username || `#${application.userId}`}
                      </span>
                      <div className="flex gap-2">
                        <Button
                          variant="outline"
                          disabled={Boolean(reviewingId)}
                          onClick={() =>
                            void review(group.id, application.id, false)
                          }
                        >
                          {t("reject")}
                        </Button>
                        <Button
                          disabled={Boolean(reviewingId)}
                          onClick={() =>
                            void review(group.id, application.id, true)
                          }
                        >
                          {reviewingId === application.id ? (
                            <Spinner data-icon="inline-start" />
                          ) : null}
                          {t("approve")}
                        </Button>
                      </div>
                    </div>
                  ))
                ) : (
                  <p className="mt-2 text-sm text-muted-foreground">
                    {t("noPendingApplications")}
                  </p>
                )}
              </article>
            ))}
          </div>
        </section>
      ) : null}
    </main>
  );
}
