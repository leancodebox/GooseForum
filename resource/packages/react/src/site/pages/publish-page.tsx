import { useEffect, useMemo, useRef, useState } from "react";
import type {
  PublishCategoryPayload,
  PublishPageProps,
} from "@gooseforum/client";
import { Check, ListChecks, Lock, Send, X } from "lucide-react";
import { useTranslation } from "react-i18next";
import { Alert, AlertDescription } from "../../components/ui/alert";
import { Badge } from "../../components/ui/badge";
import { Button } from "../../components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "../../components/ui/dialog";
import {
  Field,
  FieldDescription,
  FieldError,
  FieldGroup,
  FieldLabel,
  FieldSet,
  FieldLegend,
} from "../../components/ui/field";
import { Input } from "../../components/ui/input";
import { Spinner } from "../../components/ui/spinner";
import { GooseLink, useGooseRuntime } from "../../runtime";
import { MarkdownComposer } from "../editor/markdown-composer";
import { PageHeader } from "../layout/page-header";

export function PublishPageView({ page }: { page: PublishPageProps }) {
  const { t } = useTranslation("publish");
  const runtime = useGooseRuntime();
  const [title, setTitle] = useState(page.topic.title || "");
  const [content, setContent] = useState(page.topic.content || "");
  const [categoryIds, setCategoryIds] = useState(page.topic.categoryIds || []);
  const [topicId, setTopicId] = useState(page.topicId);
  const [submitting, setSubmitting] = useState(false);
  const [validation, setValidation] = useState(false);
  const [error, setError] = useState("");
  const [message, setMessage] = useState("");
  const [pendingHref, setPendingHref] = useState("");
  const [confirmMainCategory, setConfirmMainCategory] = useState(false);
  const mainCategoryConfirmed = useRef(false);
  const leaveResolver = useRef<((allow: boolean) => void) | null>(null);
  const allowNavigation = useRef(false);
  const initialMain = useRef(page.topic.categoryIds?.[0] || 0);
  const savedSnapshot = useRef(snapshot(title, content, categoryIds));
  const currentSnapshot = snapshot(title, content, categoryIds);
  const hasChanges = currentSnapshot !== savedSnapshot.current;
  const hasChangesRef = useRef(hasChanges);
  hasChangesRef.current = hasChanges;
  const selected = useMemo(
    () =>
      page.categories.filter((category) => categoryIds.includes(category.id)),
    [categoryIds, page.categories],
  );
  const restricted = selected.find((category) => category.isRestricted);
  const limit = restricted ? 1 : 3;
  const valid = Boolean(title.trim() && content.trim() && categoryIds.length);
  const canPublish =
    page.topic.topicStatus === 1 ||
    selected.every((category) => category.canCreate);

  useEffect(() => {
    function beforeUnload(event: BeforeUnloadEvent) {
      if (!allowNavigation.current && hasChangesRef.current) {
        event.preventDefault();
        event.returnValue = "";
      }
    }
    window.addEventListener("beforeunload", beforeUnload);
    const remove = runtime.registerNavigationBlocker?.((href) => {
      if (allowNavigation.current || !hasChangesRef.current) return true;
      setPendingHref(href);
      return new Promise<boolean>((resolve) => {
        leaveResolver.current?.(false);
        leaveResolver.current = resolve;
      });
    });
    return () => {
      window.removeEventListener("beforeunload", beforeUnload);
      remove?.();
      leaveResolver.current?.(true);
    };
  }, [runtime.registerNavigationBlocker]);

  function chooseCategory(category: PublishCategoryPayload) {
    setValidation(false);
    setCategoryIds((current) => {
      if (current.includes(category.id))
        return current.filter((id) => id !== category.id);
      if (!category.canCreate) return current;
      const currentRestricted = page.categories.some(
        (item) => current.includes(item.id) && item.isRestricted,
      );
      if (category.isRestricted || currentRestricted) return [category.id];
      return current.length >= 3 ? current : [...current, category.id];
    });
  }
  function categoryDisabled(category: PublishCategoryPayload) {
    if (categoryIds.includes(category.id)) return false;
    if (!category.canCreate) return true;
    return Boolean(
      (restricted && !category.isRestricted) ||
        (!category.isRestricted && categoryIds.length >= 3),
    );
  }
  function validate() {
    setValidation(true);
    return valid;
  }
  async function persist(status: 0 | 1, destination: string) {
    if (submitting || !validate()) return false;
    if (status === 1 && !canPublish) {
      setError(t("noCreatePermission"));
      return false;
    }
    if (
      status === 1 &&
      page.topic.topicStatus === 1 &&
      initialMain.current !== categoryIds[0] &&
      !mainCategoryConfirmed.current
    ) {
      setConfirmMainCategory(true);
      return false;
    }
    setSubmitting(true);
    setError("");
    setMessage("");
    try {
      const result = await runtime.api.topics.writeReviewed({
        topicId,
        title: title.trim(),
        content: content.trim(),
        categoryId: categoryIds,
        topicStatus: status,
      });
      setTopicId(result.id);
      savedSnapshot.current = snapshot(title, content, categoryIds);
      allowNavigation.current = true;
      if (status === 1 && result.moderationStatus === "rejected") {
        setError(t("moderationRejected"));
        allowNavigation.current = false;
        return false;
      }
      setMessage(
        status === 0
          ? t("saveDraft")
          : t(page.isEditing ? "topicUpdated" : "topicPublished"),
      );
      if (destination) {
        await runtime.navigate(destination.replace(":id", String(result.id)));
      }
      return true;
    } catch (reason) {
      setError(
        reason instanceof Error
          ? reason.message
          : t(status === 0 ? "draftSaveFailed" : "saveFailed"),
      );
      return false;
    } finally {
      setSubmitting(false);
    }
  }
  function resolveLeave(allow: boolean) {
    leaveResolver.current?.(allow);
    leaveResolver.current = null;
    setPendingHref("");
    if (allow) allowNavigation.current = true;
  }
  async function saveBeforeLeave() {
    if (!(await persist(0, ""))) return;
    resolveLeave(true);
  }

  return (
    <main className="min-w-0 pb-8">
      <PageHeader
        title={t(page.isEditing ? "editTitle" : "createTitle")}
        description={t("subtitle")}
      />
      <div className="grid gap-3 xl:grid-cols-[minmax(0,1fr)_280px]">
        <section className="rounded-xl border bg-background p-4 sm:p-5">
          <FieldGroup>
            <Field data-invalid={validation && !title.trim()}>
              <FieldLabel htmlFor="topic-title">{t("fields.title")}</FieldLabel>
              <Input
                id="topic-title"
                value={title}
                onChange={(event) => {
                  setTitle(event.target.value);
                  setValidation(false);
                }}
                aria-invalid={validation && !title.trim()}
                className="h-11 text-lg font-semibold"
                placeholder={t("titlePlaceholder")}
              />
              <FieldError>
                {validation && !title.trim()
                  ? t("validation.requiredFields")
                  : null}
              </FieldError>
            </Field>
            <FieldSet data-invalid={validation && !categoryIds.length}>
              <FieldLegend variant="label">{t("fields.category")}</FieldLegend>
              <FieldDescription>
                {categoryIds[0]
                  ? t("mainCategoryHint", {
                      category:
                        page.categories.find(
                          (category) => category.id === categoryIds[0],
                        )?.name || "",
                    })
                  : t("maxCategories")}
              </FieldDescription>
              <div className="flex flex-wrap gap-2">
                {page.categories.map((category) => (
                  <Button
                    key={category.id}
                    type="button"
                    variant={
                      categoryIds.includes(category.id)
                        ? "secondary"
                        : "outline"
                    }
                    disabled={categoryDisabled(category)}
                    aria-pressed={categoryIds.includes(category.id)}
                    title={
                      category.isRestricted
                        ? t("restrictedCategorySingleHint")
                        : undefined
                    }
                    onClick={() => chooseCategory(category)}
                  >
                    <span
                      className="size-2 rounded-sm"
                      style={{ backgroundColor: category.color }}
                    />
                    {category.name}
                    {categoryIds[0] === category.id ? (
                      <Badge variant="outline" className="px-1 text-[10px]">
                        {t("mainCategoryBadge")}
                      </Badge>
                    ) : null}
                    {category.isRestricted ? <Lock /> : null}
                  </Button>
                ))}
              </div>
              <FieldError>
                {validation && !categoryIds.length
                  ? t("validation.categoryRequired")
                  : null}
              </FieldError>
            </FieldSet>
            <Field data-invalid={validation && !content.trim()}>
              <FieldLabel>{t("fields.body")}</FieldLabel>
              <MarkdownComposer
                value={content}
                onChange={(value) => {
                  setContent(value);
                  setValidation(false);
                }}
              />
              <FieldError>
                {validation && !content.trim()
                  ? t("validation.requiredFields")
                  : null}
              </FieldError>
            </Field>
            {error ? (
              <Alert variant="destructive">
                <AlertDescription>{error}</AlertDescription>
              </Alert>
            ) : null}
            {message ? (
              <Alert>
                <Check />
                <AlertDescription>{message}</AlertDescription>
              </Alert>
            ) : null}
            <div className="flex flex-wrap justify-end gap-2 border-t pt-4">
              <Button asChild variant="ghost">
                <GooseLink href="/">{t("cancel")}</GooseLink>
              </Button>
              <Button
                type="button"
                variant="outline"
                disabled={submitting}
                onClick={() => void persist(0, "/drafts")}
              >
                {submitting ? <Spinner data-icon="inline-start" /> : null}
                {submitting ? t("saving") : t("saveDraft")}
              </Button>
              <Button
                type="button"
                disabled={submitting || !canPublish}
                onClick={() => void persist(1, "/p/post/:id")}
              >
                <Send data-icon="inline-start" />
                {submitting
                  ? t("saving")
                  : t(page.isEditing ? "updateTopic" : "publishTopic")}
              </Button>
            </div>
          </FieldGroup>
        </section>
        <aside className="flex flex-col gap-3">
          <section className="rounded-xl border bg-background p-4">
            <div className="flex items-center gap-2">
              <ListChecks />
              <h2 className="text-sm font-semibold">{t("checklist.title")}</h2>
            </div>
            <ul className="mt-3 flex flex-col gap-2 text-sm">
              <CheckItem
                label={t("fields.title")}
                done={Boolean(title.trim())}
                value={
                  title.trim() ? t("checklist.done") : t("checklist.pending")
                }
              />
              <CheckItem
                label={t("fields.category")}
                done={Boolean(categoryIds.length)}
                value={`${categoryIds.length}/${limit}`}
              />
              <CheckItem
                label={t("fields.body")}
                done={Boolean(content.trim())}
                value={t("checklist.characters", {
                  count: content.trim().length,
                })}
              />
            </ul>
          </section>
          {selected.length ? (
            <section className="rounded-xl border bg-background p-4">
              <h2 className="text-sm font-semibold">
                {t("selectedCategories")}
              </h2>
              <div className="mt-3 flex flex-wrap gap-2">
                {selected.map((category) => (
                  <Button
                    key={category.id}
                    type="button"
                    variant="outline"
                    size="sm"
                    onClick={() => chooseCategory(category)}
                  >
                    <span
                      className="size-2 rounded-sm"
                      style={{ backgroundColor: category.color }}
                    />
                    {category.name}
                    <X />
                  </Button>
                ))}
              </div>
            </section>
          ) : null}
        </aside>
      </div>
      <Dialog
        open={Boolean(pendingHref)}
        onOpenChange={(open) => {
          if (!open) resolveLeave(false);
        }}
      >
        <DialogContent showCloseButton={false}>
          <DialogHeader>
            <DialogTitle>{t("leaveTitle")}</DialogTitle>
            <DialogDescription>{t("leaveDescription")}</DialogDescription>
          </DialogHeader>
          {!valid ? (
            <Alert>
              <AlertDescription>{t("draftRequirement")}</AlertDescription>
            </Alert>
          ) : null}
          <DialogFooter>
            <Button variant="ghost" onClick={() => resolveLeave(false)}>
              {t("continueEditing")}
            </Button>
            <Button variant="outline" onClick={() => resolveLeave(true)}>
              {t("leaveWithoutSaving")}
            </Button>
            <Button
              disabled={!valid || submitting}
              onClick={() => void saveBeforeLeave()}
            >
              {submitting ? <Spinner data-icon="inline-start" /> : null}
              {t("saveDraft")}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
      <Dialog open={confirmMainCategory} onOpenChange={setConfirmMainCategory}>
        <DialogContent showCloseButton={false}>
          <DialogHeader>
            <DialogTitle>{t("mainCategoryChangeTitle")}</DialogTitle>
            <DialogDescription>
              {t("mainCategoryChangeDescription", {
                current:
                  page.categories.find(
                    (category) => category.id === initialMain.current,
                  )?.name || "",
                next: selected[0]?.name || "",
              })}
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setConfirmMainCategory(false)}
            >
              {t("cancel")}
            </Button>
            <Button
              onClick={() => {
                mainCategoryConfirmed.current = true;
                setConfirmMainCategory(false);
                requestAnimationFrame(() => void persist(1, "/p/post/:id"));
              }}
            >
              {t("confirmMainCategoryChange")}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </main>
  );
}

function CheckItem({
  label,
  done,
  value,
}: {
  label: string;
  done: boolean;
  value: string;
}) {
  return (
    <li className="flex items-center justify-between gap-3">
      <span>{label}</span>
      <span className={done ? "text-success" : "text-muted-foreground"}>
        {value}
      </span>
    </li>
  );
}
function snapshot(title: string, content: string, categories: number[]) {
  return JSON.stringify({
    title: title.trim(),
    content: content.trim(),
    categories,
  });
}
