import { useEffect, useRef } from 'react';
import { combine } from '@atlaskit/pragmatic-drag-and-drop/combine';
import {
  dropTargetForElements,
} from '@atlaskit/pragmatic-drag-and-drop/element/adapter';
import { Plus } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { useBoardContext } from './board-context';
import { UserCard } from './user-card';
import { UserSponsor } from '../data/schema';

interface UserColumnProps {
  users: UserSponsor[];
  onAdd: () => void;
  onEdit: (index: number) => void;
  onDelete: (index: number) => void;
}

export function UserColumn({
  users,
  onAdd,
  onEdit,
  onDelete,
}: UserColumnProps) {
  const { instanceId } = useBoardContext();
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const element = ref.current;
    if (!element) return;

    return combine(
      dropTargetForElements({
        element,
        canDrop: ({ source }) =>
          source.data.instanceId === instanceId && source.data.type === 'user',
        getIsSticky: () => true,
        getData: () => ({
          type: 'user-column',
        }),
      })
    );
  }, [instanceId]);

  return (
    <div
      ref={ref}
      className="relative flex flex-col w-full bg-[#f4f5f7] rounded-2xl transition-all border-transparent"
    >
      <div
        className="flex items-center justify-between px-4 py-3 border-b border-[#091e4214] hover:bg-[#091e4208] transition-colors rounded-t-2xl"
      >
        <div className="flex items-center gap-2 overflow-hidden">
          <h3 className="text-[12px] font-bold text-[#44546f] uppercase tracking-tight truncate">
            个人赞助者
          </h3>
          <span className="flex-shrink-0 min-w-[20px] h-5 flex items-center justify-center bg-[#091e420f] text-[#44546f] text-[11px] font-medium px-1.5 rounded-full">
            {users.length}
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

      <div className="p-3 grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6 xl:grid-cols-8 gap-2 min-h-[120px]">
        {users.map((user, index) => (
          <UserCard
            key={`user-${user.name}-${index}`}
            user={user}
            index={index}
            onEdit={() => onEdit(index)}
            onDelete={() => onDelete(index)}
          />
        ))}
      </div>
    </div>
  );
}
