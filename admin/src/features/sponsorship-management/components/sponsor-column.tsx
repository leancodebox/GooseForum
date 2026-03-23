import { useEffect, useRef, useState } from 'react';
import { combine } from '@atlaskit/pragmatic-drag-and-drop/combine';
import {
  dropTargetForElements,
} from '@atlaskit/pragmatic-drag-and-drop/element/adapter';
import {
  attachClosestEdge,
  extractClosestEdge,
} from '@atlaskit/pragmatic-drag-and-drop-hitbox/closest-edge';
import type { Edge } from '@atlaskit/pragmatic-drag-and-drop-hitbox/types';
import { Plus } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { useBoardContext } from './board-context';
import { SponsorCard } from './sponsor-card';
import { DropIndicator } from './drop-indicator';
import { SponsorItem, Sponsors } from '../data/schema';

interface SponsorColumnProps {
  level: keyof Sponsors;
  title: string;
  sponsors: SponsorItem[];
  onAdd: () => void;
  onEdit: (index: number) => void;
  onDelete: (index: number) => void;
}

export function SponsorColumn({
  level,
  title,
  sponsors,
  onAdd,
  onEdit,
  onDelete,
}: SponsorColumnProps) {
  const { registerLevel, instanceId } = useBoardContext();
  const ref = useRef<HTMLDivElement>(null);
  const headerRef = useRef<HTMLDivElement>(null);
  const [closestEdge, setClosestEdge] = useState<Edge | null>(null);

  useEffect(() => {
    const element = ref.current;
    const header = headerRef.current;
    if (!element || !header) return;

    // We don't necessarily need columns to be draggable for now, 
    // but we'll implement it to match the links-management style if needed.
    // For now, let's just make them drop targets for sponsors.
    return combine(
      registerLevel(level, { element }),
      dropTargetForElements({
        element,
        canDrop: ({ source }) =>
          source.data.instanceId === instanceId && source.data.type === 'sponsor',
        getIsSticky: () => true,
        getData: ({ input, element }) => {
          const data = {
            type: 'level',
            level,
          };
          return attachClosestEdge(data, {
            input,
            element,
            allowedEdges: ['top', 'bottom'],
          });
        },
        onDragEnter: ({ self, source }) => {
          if (source.data.type === 'level') {
            setClosestEdge(extractClosestEdge(self.data));
          }
        },
        onDrag: ({ self, source }) => {
          if (source.data.type === 'level') {
            setClosestEdge(extractClosestEdge(self.data));
          }
        },
        onDragLeave: () => setClosestEdge(null),
        onDrop: () => setClosestEdge(null),
      })
    );
  }, [level, instanceId, registerLevel]);

  return (
    <div
      ref={ref}
      className={`relative flex flex-col w-full bg-[#f4f5f7] rounded-2xl transition-all border-transparent`}
    >
      {closestEdge && <DropIndicator edge={closestEdge} />}
      <div
        ref={headerRef}
        className="flex items-center justify-between px-4 py-3 border-b border-[#091e4214] hover:bg-[#091e4208] transition-colors rounded-t-2xl"
      >
        <div className="flex items-center gap-2 overflow-hidden">
          <h3 className="text-[12px] font-bold text-[#44546f] uppercase tracking-tight truncate">
            {title}
          </h3>
          <span className="flex-shrink-0 min-w-[20px] h-5 flex items-center justify-center bg-[#091e420f] text-[#44546f] text-[11px] font-medium px-1.5 rounded-full">
            {sponsors.length}
          </span>
        </div>
        <Button
          variant="ghost"
          size="sm"
          className="h-7 px-2 text-[#44546f] hover:bg-[#091e420f] text-[12px] font-medium"
          onClick={onAdd}
        >
          <Plus className="mr-1 h-3.5 w-3.5" />
          添加
        </Button>
      </div>

      <div className="p-3 grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 gap-3 min-h-[120px]">
        {sponsors.map((sponsor, index) => (
          <SponsorCard
            key={`${level}-${sponsor.name}-${index}`}
            sponsor={sponsor}
            level={level}
            index={index}
            onEdit={() => onEdit(index)}
            onDelete={() => onDelete(index)}
          />
        ))}
      </div>
    </div>
  );
}
