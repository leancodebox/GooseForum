import { useEffect, useRef, useState } from 'react';
import { combine } from '@atlaskit/pragmatic-drag-and-drop/combine';
import {
  draggable,
  dropTargetForElements,
} from '@atlaskit/pragmatic-drag-and-drop/element/adapter';
import {
  attachClosestEdge,
  extractClosestEdge,
} from '@atlaskit/pragmatic-drag-and-drop-hitbox/closest-edge';
import type { Edge } from '@atlaskit/pragmatic-drag-and-drop-hitbox/types';
import { ExternalLink, Eye, EyeOff, GripVertical, Link as LinkIcon, MoreHorizontal } from 'lucide-react';
import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { useBoardContext } from './board-context';
import { DropIndicator } from './drop-indicator';
import { Link } from '../data/schema';
import { cn } from '@/lib/utils';

interface LinkCardProps {
  link: Link;
  groupIdx: number;
  linkIdx: number;
  onEdit: () => void;
  onDelete: () => void;
  onToggleStatus: () => void;
}

export function LinkCard({ link, groupIdx, linkIdx, onEdit, onDelete, onToggleStatus }: LinkCardProps) {
  const { registerLink, instanceId } = useBoardContext();
  const ref = useRef<HTMLDivElement>(null);
  const [dragging, setDragging] = useState(false);
  const [closestEdge, setClosestEdge] = useState<Edge | null>(null);

  const linkId = `${groupIdx}-${linkIdx}-${link.name}`;

  useEffect(() => {
    const element = ref.current;
    if (!element) return;

    return combine(
      registerLink(linkId, { element, actionMenuTrigger: element }),
      draggable({
        element,
        getInitialData: () => ({
          type: 'link',
          linkId,
          groupIdx,
          linkIdx,
          instanceId,
        }),
        onDragStart: () => setDragging(true),
        onDrop: () => setDragging(false),
      }),
      dropTargetForElements({
        element,
        canDrop: ({ source }) =>
          source.data.instanceId === instanceId && source.data.type === 'link',
        getIsSticky: () => true,
        getData: ({ input, element }) => {
          const data = {
            type: 'link',
            linkId,
            groupIdx,
            linkIdx,
          };
          return attachClosestEdge(data, {
            input,
            element,
            allowedEdges: ['left', 'right'],
          });
        },
        onDragEnter: ({ self }) => setClosestEdge(extractClosestEdge(self.data)),
        onDrag: ({ self }) => setClosestEdge(extractClosestEdge(self.data)),
        onDragLeave: () => setClosestEdge(null),
        onDrop: () => setClosestEdge(null),
      })
    );
  }, [linkId, groupIdx, linkIdx, instanceId, registerLink]);

  return (
    <div
      ref={ref}
      className={cn(
        'group relative rounded-md border bg-background px-2.5 py-2 transition hover:border-primary/30 hover:bg-primary/[0.02]',
        dragging && 'opacity-50 grayscale',
        link.status === 0 && 'bg-muted/40 opacity-65'
      )}
    >
      {closestEdge && <DropIndicator edge={closestEdge} />}
      <GripVertical className="absolute left-1 top-1 h-3.5 w-3.5 cursor-grab text-muted-foreground/30 opacity-0 transition-opacity group-hover:opacity-100" />

      <div className="flex items-center gap-2">
        <div className="relative shrink-0">
          <Avatar className={cn('h-8 w-8 rounded-md border border-border/70 bg-muted', link.status === 0 && 'grayscale')}>
            <AvatarImage src={link.logoUrl} alt={link.name} className="object-cover" />
            <AvatarFallback className="rounded-md bg-muted text-muted-foreground">
              <LinkIcon className="h-4 w-4" />
            </AvatarFallback>
          </Avatar>
          <button
            onClick={(e) => {
              e.stopPropagation();
              onToggleStatus();
            }}
            className={cn(
              'absolute -bottom-1 -right-1 rounded-full border p-0.5 shadow-sm transition-all',
              link.status === 0
                ? 'bg-background text-muted-foreground hover:bg-muted'
                : 'bg-primary text-primary-foreground opacity-0 group-hover:opacity-100'
            )}
            title={link.status === 0 ? '点击展示链接' : '点击隐藏链接'}
          >
            {link.status === 0 ? (
              <EyeOff className="h-3 w-3" />
            ) : (
              <Eye className="h-3 w-3" />
            )}
          </button>
        </div>

        <div className="min-w-0 flex-1">
          <div className="flex min-w-0 items-center gap-2">
            <h4 className={cn('truncate text-[13px] font-semibold text-foreground group-hover:text-primary', link.status === 0 && 'text-muted-foreground')}>
              {link.name}
            </h4>
            <a
              href={link.url}
              target="_blank"
              rel="noopener noreferrer"
              className="shrink-0 text-muted-foreground/40 transition-colors hover:text-primary"
              aria-label="访问链接"
              onClick={(event) => event.stopPropagation()}
            >
              <ExternalLink className="h-3 w-3" />
            </a>
          </div>
          <p className="mt-0.5 truncate text-[11px] leading-4 text-muted-foreground">
            {link.desc || link.url}
          </p>
        </div>
      </div>

      <div className="absolute right-1 top-1 opacity-0 transition-opacity group-hover:opacity-100">
        <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button
            variant="ghost"
            size="icon"
            className="h-7 w-7 rounded-md bg-background/90 shadow-sm"
          >
            <MoreHorizontal className="h-4 w-4" />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end">
          <DropdownMenuItem onClick={onEdit}>编辑链接</DropdownMenuItem>
          <DropdownMenuItem className="text-destructive" onClick={onDelete}>
            删除链接
          </DropdownMenuItem>
          <DropdownMenuItem asChild>
            <a href={link.url} target="_blank" rel="noopener noreferrer">
              访问链接
            </a>
          </DropdownMenuItem>
        </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </div>
  );
}
