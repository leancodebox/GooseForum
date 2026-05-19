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
import { GripVertical, Link as LinkIcon, MoreHorizontal, Plus } from 'lucide-react';
import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { useBoardContext } from './board-context';
import { LinkCard } from './link-card';
import { DropIndicator } from './drop-indicator';
import { LinkGroup } from '../data/schema';
import { cn } from '@/lib/utils';

interface ColumnProps {
  group: LinkGroup;
  groupIdx: number;
  onEditGroup: () => void;
  onDeleteGroup: () => void;
  onAddLink: () => void;
  onEditLink: (linkIdx: number) => void;
  onDeleteLink: (linkIdx: number) => void;
  onToggleLinkStatus: (linkIdx: number) => void;
}

export function Column({
  group,
  groupIdx,
  onEditGroup,
  onDeleteGroup,
  onAddLink,
  onEditLink,
  onDeleteLink,
  onToggleLinkStatus,
}: ColumnProps) {
  const { registerGroup, instanceId } = useBoardContext();
  const ref = useRef<HTMLDivElement>(null);
  const headerRef = useRef<HTMLDivElement>(null);
  const [dragging, setDragging] = useState(false);
  const [closestEdge, setClosestEdge] = useState<Edge | null>(null);

  useEffect(() => {
    const element = ref.current;
    const header = headerRef.current;
    if (!element || !header) return;

    return combine(
      registerGroup(group.name, { element }),
      draggable({
        element: header,
        getInitialData: () => ({
          type: 'group',
          groupId: group.name,
          groupIdx,
          instanceId,
        }),
        onDragStart: () => setDragging(true),
        onDrop: () => setDragging(false),
      }),
      dropTargetForElements({
        element,
        canDrop: ({ source }) =>
          source.data.instanceId === instanceId && (source.data.type === 'link' || source.data.type === 'group'),
        getIsSticky: () => true,
        getData: ({ input, element }) => {
          const data = {
            type: 'group',
            groupId: group.name,
            groupIdx,
          };
          return attachClosestEdge(data, {
            input,
            element,
            allowedEdges: ['top', 'bottom'],
          });
        },
        onDragEnter: ({ self, source }) => {
          if (source.data.type === 'group') {
            setClosestEdge(extractClosestEdge(self.data));
          }
        },
        onDrag: ({ self, source }) => {
          if (source.data.type === 'group') {
            setClosestEdge(extractClosestEdge(self.data));
          }
        },
        onDragLeave: () => setClosestEdge(null),
        onDrop: () => setClosestEdge(null),
      })
    );
  }, [group.name, groupIdx, instanceId, registerGroup]);

  return (
    <div
      ref={ref}
      className={cn('relative space-y-3 transition-all', dragging && 'opacity-50 grayscale')}
    >
      {closestEdge && <DropIndicator edge={closestEdge} />}
      <div
        ref={headerRef}
        className="group/header flex cursor-grab items-center justify-between gap-3 border-b border-border/70 pb-2 active:cursor-grabbing"
      >
        <div className="flex min-w-0 items-center gap-2">
          <GripVertical className="h-4 w-4 shrink-0 text-muted-foreground/30 opacity-0 transition-opacity group-hover/header:opacity-100" />
          <span
            className="flex h-7 w-7 shrink-0 items-center justify-center rounded-md bg-muted text-sm"
            style={{ color: group.color || '#64748b' }}
          >
            {group.emoji || <LinkIcon className="h-4 w-4" />}
          </span>
          <h3 className="truncate text-base font-semibold text-foreground">
            {group.name}
          </h3>
          <span className="rounded-full bg-muted px-2 py-0.5 text-[11px] font-semibold text-muted-foreground">
            {group.links.length}
          </span>
        </div>
        <div className="flex shrink-0 items-center gap-1">
          <Button
            variant="ghost"
            size="sm"
            className="h-8 gap-1.5 px-2 text-xs"
            onClick={onAddLink}
          >
            <Plus className="h-3.5 w-3.5" />
            添加
          </Button>
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" size="icon" className="h-8 w-8 rounded-md">
                <MoreHorizontal className="h-4 w-4" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuItem onClick={onEditGroup}>编辑分组</DropdownMenuItem>
              <DropdownMenuItem className="text-destructive" onClick={onDeleteGroup}>
                删除分组
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </div>

      <div
        className={cn(
          'min-h-24 rounded-lg border border-dashed border-transparent',
          group.links.length === 0 && 'border-muted-foreground/20 bg-muted/20'
        )}
      >
        {group.links.length === 0 ? (
          <button
            type="button"
            onClick={onAddLink}
            className="flex h-28 w-full flex-col items-center justify-center rounded-lg text-sm text-muted-foreground transition-colors hover:bg-muted/50 hover:text-foreground"
          >
            <LinkIcon className="mb-2 h-5 w-5" />
            添加这个分组的第一个链接
          </button>
        ) : (
          <div className="grid grid-cols-2 gap-2 md:grid-cols-3 lg:grid-cols-4 2xl:grid-cols-5">
            {group.links.map((link, lIdx) => (
              <LinkCard
                key={`${group.name}-${link.name}-${lIdx}`}
                link={link}
                groupIdx={groupIdx}
                linkIdx={lIdx}
                onEdit={() => onEditLink(lIdx)}
                onDelete={() => onDeleteLink(lIdx)}
                onToggleStatus={() => onToggleLinkStatus(lIdx)}
              />
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
