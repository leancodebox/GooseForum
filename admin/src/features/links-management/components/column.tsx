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
import { Plus, MoreHorizontal } from 'lucide-react';
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
      className={`relative flex flex-col w-full bg-[#f4f5f7] rounded-2xl transition-all ${
        dragging ? 'opacity-50 grayscale' : ''
      } border-transparent`}
      style={{ 
        borderTop: group.color ? `4px solid ${group.color}` : 'none' 
      }}
    >
      {closestEdge && <DropIndicator edge={closestEdge} />}
      <div
        ref={headerRef}
        className="flex items-center justify-between px-4 py-3 cursor-grab active:cursor-grabbing border-b border-[#091e4214] hover:bg-[#091e4208] transition-colors rounded-t-2xl"
      >
        <div className="flex items-center gap-3 overflow-hidden">
          {group.emoji ? (
            <span className="text-xl flex-shrink-0 drop-shadow-sm">{group.emoji}</span>
          ) : (
            <span className="text-xl flex-shrink-0 drop-shadow-sm">🔗</span>
          )}
          <h3 
            className="text-[14px] font-bold uppercase tracking-wide truncate"
            style={{ color: group.color || '#44546f' }}
          >
            {group.name}
          </h3>
          <span className="flex-shrink-0 min-w-[22px] h-5 flex items-center justify-center bg-white/80 border border-[#091e4214] text-[#44546f] text-[11px] font-bold px-1.5 rounded-full shadow-sm">
            {group.links.length}
          </span>
        </div>
        <div className="flex items-center gap-1">
          <Button
            variant="ghost"
            size="sm"
            className="h-7 px-2 text-[#44546f] hover:bg-[#091e420f] text-[12px] font-medium"
            onClick={onAddLink}
          >
            <Plus className="mr-1 h-3.5 w-3.5" />
            添加
          </Button>
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" size="icon" className="h-7 w-7 text-[#44546f] hover:bg-[#091e420f]">
                <MoreHorizontal className="h-3.5 w-3.5" />
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

      <div className="p-4 grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 2xl:grid-cols-6 gap-4 min-h-[100px]">
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
    </div>
  );
}

