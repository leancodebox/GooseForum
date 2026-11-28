import { useEffect, useRef, useState } from "react";
import { Check, Download, Link, Share2, X, type LucideIcon } from "lucide-react";
import { QRCodeCanvas } from "qrcode.react";
import { useTranslation } from "react-i18next";
import { avatarInitial } from "@gooseforum/ui/lib/avatar-initial";
import { DefaultBrand } from "../layout/default-brand";
import type { LayoutPayload, PostPayload } from "@gooseforum/client";
import { Button } from "@gooseforum/ui/components/button";
import { Dialog, DialogContent, DialogTitle, DialogTrigger } from "@gooseforum/ui/components/dialog";
import { Input } from "@gooseforum/ui/components/input";
import { userDisplayName } from "../users/display-name";

interface ShareProps {
  post: PostPayload;
  topic: { title: string; description: string };
  site: LayoutPayload["site"];
  path: string;
}

export function PostShare(props: ShareProps) {
  const { t } = useTranslation("site.topic");
  const [open, setOpen] = useState(false);
  const first = props.post.postNo === 1;
  const label = t(first ? "shareTopic" : "shareReply");
  return <Dialog open={open} onOpenChange={setOpen}>
    <DialogTrigger asChild>
      <Button variant="ghost" size="icon-xs" aria-label={label} title={label}>
        <Share2 />
      </Button>
    </DialogTrigger>
    <DialogContent showCloseButton={false} aria-describedby={undefined}
      overlayClassName="bg-background/35 supports-backdrop-filter:backdrop-blur-[2px]"
      className="flex w-[calc(100vw-2rem)] max-w-sm flex-col gap-3 overflow-visible rounded-none bg-transparent p-0 text-foreground ring-0 sm:max-w-sm">
      <DialogTitle className="sr-only">{label}</DialogTitle>
      <CardPreview onClose={() => setOpen(false)} {...props} url={new URL(props.path, window.location.origin).href} />
    </DialogContent>
  </Dialog>;
}

function wrapLines(ctx: CanvasRenderingContext2D, text: string, width: number, limit: number) {
  const lines: string[] = [];
  let line: string[] = [];
  const graphemes = typeof Intl.Segmenter === "function"
    ? Array.from(new Intl.Segmenter().segment(text.trim()), part => part.segment)
    : Array.from(text.trim());
  // Keep names such as Next.js and react-share together when they fit a line.
  const tokens: string[] = [];
  for (const glyph of graphemes) {
    if (/^[A-Za-z0-9_.-]$/.test(glyph) && /^[A-Za-z0-9_.-]+$/.test(tokens.at(-1) || "")) {
      tokens[tokens.length - 1] += glyph;
    } else {
      tokens.push(glyph);
    }
  }
  const parts = tokens.flatMap(token => ctx.measureText(token).width > width ? Array.from(token) : [token]);
  for (const part of parts) {
    if (!line.length && part === " ") continue;
    if (part === "\n" || ctx.measureText(line.join("") + part).width > width) {
      if (lines.length === limit - 1) {
        while (line.length && ctx.measureText(line.join("") + "…").width > width) line.pop();
        return [...lines, line.join("") + "…"];
      }
      lines.push(line.join("").trimEnd());
      line = [];
      if (part === "\n" || part === " ") continue;
    }
    line.push(part);
  }
  if (line.length) lines.push(line.join("").trimEnd());
  return lines;
}

function shareExcerpt(body: HTMLElement) {
  let text = "";
  function endBlock() {
    text = text.trimEnd() + "\n";
  }
  function visit(node: Node) {
    if (node.nodeType === Node.TEXT_NODE) {
      text += (node.textContent || "").replace(/[\t\r\n ]+/g, " ");
      return;
    }
    if (!(node instanceof Element)) return;
    if (node.tagName === "BR") {
      text += "\n";
      return;
    }
    const block = node.matches("p,li,ul,ol,blockquote,pre,h1,h2,h3,h4,h5,h6,tr,div");
    if (block) endBlock();
    if (node.tagName === "PRE") text += node.textContent;
    else node.childNodes.forEach(visit);
    if (block) endBlock();
  }
  body.childNodes.forEach(visit);
  return text.replace(/ +\n/g, "\n").trim();
}

function CardPreview({ post, topic, site, url, onClose }: ShareProps & { url: string; onClose(): void }) {
  const { t } = useTranslation("site.topic");
  const brand = useRef<HTMLSpanElement>(null);
  const fallbackTone = useRef<HTMLSpanElement>(null);
  const qr = useRef<HTMLCanvasElement>(null);
  const canvas = useRef<HTMLCanvasElement>(null);
  const [file, setFile] = useState<File | null>();
  const [copyStatus, setCopyStatus] = useState<"linkCopied" | "copyManually">();
  const filename = `topic-${post.topicId}-${post.postNo}.png`;

  useEffect(() => {
    const el = canvas.current;
    const ctx = el?.getContext("2d");
    if (!el || !ctx || !qr.current) { setFile(null); return; }
    let active = true;
    const first = post.postNo === 1;
    const doc = new DOMParser().parseFromString(post.renderedContent, "text/html");
    const excerpt = shareExcerpt(doc.body) || (first ? topic.description : "");
    const author = userDisplayName(post.author);
    const palette = getComputedStyle(el);
    const paper = palette.backgroundColor;
    const ink = palette.color;
    const mutedInk = palette.getPropertyValue("--gf-color-icon-muted").trim() || ink;
    const accent = brand.current ? getComputedStyle(brand.current).color : "#2563eb";
    function draw(cover?: HTMLImageElement, logo?: HTMLImageElement, avatar?: HTMLImageElement) {
      if (!active || !ctx || !el || !qr.current) return;
      const bannerHeight = cover ? 134 : 64;
      ctx.font = "550 20px system-ui";
      const heading = wrapLines(ctx, topic.title, 304, 3);
      const top = bannerHeight + 42;
      const quoteTop = top + (heading.length - 1) * 27 + 28;
      ctx.font = "400 15px system-ui";
      const detail = wrapLines(ctx, excerpt, 288, 8);
      const contentBottom = detail.length
        ? quoteTop - 15 + detail.length * 23
        : top + (heading.length - 1) * 27 + 5;
      const footerY = contentBottom + 22;
      const height = footerY + 100;
      el.height = height * 2;
      ctx.setTransform(2, 0, 0, 2, 0, 0);
      ctx.fillStyle = paper;
      ctx.fillRect(0, 0, 360, height);
      const gradient = ctx.createLinearGradient(0, 0, 360, bannerHeight);
      gradient.addColorStop(0, accent);
      gradient.addColorStop(1, paper);
      ctx.globalAlpha = 0.18;
      ctx.fillStyle = gradient;
      ctx.fillRect(0, 0, 360, bannerHeight);
      ctx.globalAlpha = 1;
      if (cover) {
        const scale = Math.max(360 / cover.naturalWidth, bannerHeight / cover.naturalHeight);
        const w = 360 / scale, h = bannerHeight / scale;
        ctx.drawImage(cover, (cover.naturalWidth-w)/2, (cover.naturalHeight-h)/2, w, h, 0, 0, 360, bannerHeight);
        ctx.fillStyle = paper;
        ctx.globalAlpha = 0.9;
        ctx.fillRect(0, 0, 360, 48);
        ctx.globalAlpha = 1;
      }
      const baseline = cover ? 31 : 45;
      ctx.fillStyle = ink;
      ctx.font = "700 19px system-ui";
      if (logo) {
        const scale = Math.min(180 / logo.naturalWidth, 34 / logo.naturalHeight);
        ctx.drawImage(logo, 28, (cover ? 48 : bannerHeight) / 2 - logo.naturalHeight * scale / 2, logo.naturalWidth * scale, logo.naturalHeight * scale);
      } else {
        ctx.fillText(wrapLines(ctx, site.brandText || site.name, 192, 1)[0] || "", 28, baseline);
      }
      ctx.font = "13px system-ui";
      ctx.textAlign = "right";
      ctx.fillText(post.createdAt.slice(0, 10), 332, baseline);
      ctx.textAlign = "left";
      ctx.font = "550 20px system-ui";
      heading.forEach((line, i) => ctx.fillText(line, 28, top + i * 27));
      ctx.font = "400 15px system-ui";
      if (detail.length) {
        ctx.save();
        ctx.fillStyle = accent;
        ctx.globalAlpha = 0.3;
        ctx.fillRect(29, quoteTop - 15, 3, detail.length * 23);
        ctx.restore();
        ctx.fillStyle = ink;
        detail.forEach((line, i) => ctx.fillText(line, 44, quoteTop + i * 23));
      }
      ctx.fillStyle = palette.borderTopColor;
      ctx.fillRect(28, footerY, 304, 1);
      ctx.save();
      ctx.fillStyle = accent;
      ctx.globalAlpha = 0.09;
      ctx.beginPath();
      ctx.arc(0, height, 58, 0, Math.PI * 2);
      ctx.fill();
      ctx.restore();
      ctx.save();
      ctx.beginPath();
      ctx.arc(58, footerY + 50, 30, 0, Math.PI * 2);
      ctx.clip();
      if (avatar) {
        const side = Math.min(avatar.naturalWidth, avatar.naturalHeight);
        ctx.drawImage(avatar, (avatar.naturalWidth - side) / 2, (avatar.naturalHeight - side) / 2, side, side, 28, footerY + 20, 60, 60);
      } else {
        const fallback = getComputedStyle(fallbackTone.current!);
        ctx.fillStyle = fallback.backgroundColor;
        ctx.fillRect(28, footerY + 20, 60, 60);
        ctx.fillStyle = fallback.color;
        ctx.font = "600 24px system-ui";
        ctx.textAlign = "center";
        ctx.fillText(avatarInitial(author), 58, footerY + 58);
      }
      ctx.restore();
      ctx.fillStyle = mutedInk;
      ctx.font = "12px system-ui";
      ctx.fillText(t(first ? "topicAuthor" : "replyAuthor"), 100, footerY + 42);
      ctx.fillStyle = ink;
      ctx.font = "550 15px system-ui";
      ctx.fillText(wrapLines(ctx, author, 158, 1)[0] || "", 100, footerY + 68);
      ctx.imageSmoothingEnabled = false;
      ctx.drawImage(qr.current, 270, footerY + 12, 62, 62);
      ctx.fillStyle = mutedInk;
      ctx.font = "10px system-ui";
      ctx.textAlign = "center";
      ctx.fillText(t("scanToRead"), 301, footerY + 86, 90);
      ctx.textAlign = "left";
      el.toBlob(blob => {
        if (!active) return;
        if (blob) setFile(new File([blob], filename, { type: "image/png" }));
        else setFile(null);
      }, "image/png");
    }

    function loadImage(src?: string | null): Promise<HTMLImageElement | undefined> {
      if (!src) return Promise.resolve(undefined);
      const image = new Image();
      image.crossOrigin = "anonymous";
      image.src = src;
      return new Promise(resolve => {
        const timeout = window.setTimeout(() => resolve(undefined), 2500);
        image.decode().then(() => resolve(image), () => resolve(undefined)).finally(() => clearTimeout(timeout));
      });
    }
    const svg = brand.current?.querySelector("svg")?.cloneNode(true) as SVGSVGElement | undefined;
    if (svg) {
      svg.style.color = ink;
      svg.style.setProperty("--primary", accent);
    }
    const logoSrc = svg ? "data:image/svg+xml;charset=utf-8," + encodeURIComponent(new XMLSerializer().serializeToString(svg)) : site.brandImage;
    Promise.all([
      loadImage(first ? doc.querySelector("img")?.getAttribute("src") : undefined),
      loadImage(site.brandType === "text" ? undefined : logoSrc),
      loadImage(post.author.avatarUrl),
    ]).then(([cover, logo, avatar]) => draw(cover, logo, avatar)).catch(() => {
      if (active) setFile(null);
    });
    return () => { active = false; };
  }, [post, topic, site, t, filename, url]);

  function save() {
    if (!file) return;
    const href = URL.createObjectURL(file);
    const link = document.createElement("a");
    link.href = href;
    link.download = filename;
    link.click();
    window.setTimeout(() => URL.revokeObjectURL(href), 1000);
  }

  async function shareImage() {
    if (!file) return;
    try { await navigator.share({ files: [file] }); }
    catch (error) { if (!(error instanceof Error && error.name === "AbortError")) save(); }
  }

  async function copyLink() {
    try {
      await navigator.clipboard.writeText(url);
      setCopyStatus("linkCopied");
    } catch {
      setCopyStatus("copyManually");
    }
  }

  const native = file && typeof navigator.share === "function" && navigator.canShare?.({ files: [file] });
  return <>
    <span ref={brand} className="hidden text-primary" aria-hidden="true">{site.brandType !== "text" && !(site.brandType === "image" && site.brandImage) && <DefaultBrand />}</span>
    <span ref={fallbackTone} className="hidden bg-muted text-muted-foreground" data-slot="avatar-fallback" data-avatar-tone={post.author.id % 8} aria-hidden="true" />
    <QRCodeCanvas ref={qr} value={url} size={240} marginSize={4} className="hidden" />
    <canvas ref={canvas} width={720} role="img" aria-label={`${topic.title} · #${post.postNo}`}
      className="mx-auto block h-auto max-h-[calc(100dvh-7rem)] w-auto max-w-full rounded-xl border-border bg-card text-card-foreground shadow-xl ring-1 ring-border md:max-h-[calc(100dvh-2rem)]" />
    {file === null && <p role="alert">{t("cardFailed")}</p>}
    {copyStatus === "copyManually" && <label>{t("copyManually")}<Input autoFocus readOnly value={url} onFocus={event => event.target.select()} /></label>}
    <div className="flex justify-center gap-5 md:absolute md:left-full md:top-1/2 md:ml-4 md:-translate-y-1/2 md:flex-col md:gap-3">
      <PreviewAction label={t(native ? "shareCard" : "saveCard")} icon={native ? Share2 : Download}
        primary disabled={!file} onClick={native ? shareImage : save} />
      <PreviewAction label={t(copyStatus === "linkCopied" ? "linkCopied" : "copyLink")} icon={copyStatus === "linkCopied" ? Check : Link} onClick={copyLink} />
      {native && <PreviewAction label={t("saveCard")} icon={Download} onClick={save} />}
      <div className="ml-1 md:ml-0 md:mt-1"><PreviewAction label={t("close")} icon={X} onClick={onClose} /></div>
    </div>
  </>;
}

function PreviewAction({ label, icon: Icon, primary, disabled, onClick }: {
  label: string; icon: LucideIcon; primary?: boolean; disabled?: boolean; onClick(): void;
}) {
  return <button type="button" aria-label={label} disabled={disabled} onClick={onClick}
    className="group relative flex flex-col items-center gap-1.5 rounded-full focus-visible:outline-2 focus-visible:outline-primary disabled:opacity-50">
    <span className={`grid size-11 place-items-center rounded-full shadow-sm transition-transform group-hover:scale-105 ${primary ? "bg-primary text-primary-foreground" : "border border-border bg-popover text-popover-foreground"}`}>
      <Icon className="size-5" />
    </span>
    <span aria-live="polite" className="text-xs font-medium md:pointer-events-none md:absolute md:right-full md:top-1/2 md:mr-2 md:-translate-y-1/2 md:rounded-md md:bg-popover md:text-popover-foreground md:px-2 md:py-1 md:whitespace-nowrap md:opacity-0 md:shadow-sm md:group-hover:opacity-100 md:group-focus-visible:opacity-100">{label}</span>
  </button>;
}
