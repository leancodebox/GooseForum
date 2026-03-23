import { useEffect, useState, useCallback, useMemo, useRef } from 'react';
import { toast } from 'sonner';
import { Plus } from 'lucide-react';
import { combine } from '@atlaskit/pragmatic-drag-and-drop/combine';
import { monitorForElements } from '@atlaskit/pragmatic-drag-and-drop/element/adapter';
import { reorder } from '@atlaskit/pragmatic-drag-and-drop/reorder';
import { extractClosestEdge } from '@atlaskit/pragmatic-drag-and-drop-hitbox/closest-edge';
import { getReorderDestinationIndex } from '@atlaskit/pragmatic-drag-and-drop-hitbox/util/get-reorder-destination-index';

import { Button } from '@/components/ui/button';
import { ContentLayout } from '@/components/layout/content-layout';
import { LinksActionDialog } from './components/links-action-dialog';
import { LinksDeleteDialog } from './components/links-delete-dialog';
import { LinksGroupActionDialog } from './components/links-group-action-dialog';
import LinksProvider, { useLinks } from './components/links-provider';
import { type Link, type LinkGroup } from './data/schema';
import { Column } from './components/column';
import { BoardContext } from './components/board-context';
import { createRegistry } from './components/registry';
import { getFriendLinks, saveFriendLinks } from '@/api';

function LinksManagementContent() {
  const {
    open,
    setOpen,
    currentRow,
    setCurrentRow,
    currentGroup,
    setCurrentGroup,
    groupIndex,
    setGroupIndex,
    linkIndex,
    setLinkIndex,
  } = useLinks();

  const [groups, setGroups] = useState<LinkGroup[]>([]);
  const groupsRef = useRef<LinkGroup[]>(groups);
  groupsRef.current = groups;
  const [loading, setLoading] = useState(true);
  const [instanceId] = useState(() => Symbol('instance-id'));
  const [registry] = useState(createRegistry);

  const fetchLinks = async () => {
    try {
      const res = await getFriendLinks();
      if (res.code === 0) {
        setGroups(res.result || []);
      } else {
        toast.error(res.msg || '获取友情链接失败');
      }
    } catch (error) {
      toast.error("获取友情链接失败");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchLinks();
  }, []);

  const saveLinks = useCallback(async (newGroups: LinkGroup[]) => {
    try {
      const res = await saveFriendLinks(newGroups);
      if (res.code === 0) {
        toast.success('保存成功');
      } else {
        toast.error(res.msg || '保存失败');
        fetchLinks();
      }
    } catch (error) {
      toast.error('保存失败');
      fetchLinks();
    }
  }, []);

  const reorderGroup = useCallback(({ startIndex, finishIndex }: { startIndex: number; finishIndex: number }) => {
    const newGroups = reorder({
      list: groupsRef.current,
      startIndex,
      finishIndex,
    });
    setGroups(newGroups);
    saveLinks(newGroups);
  }, [saveLinks]);

  const reorderLink = useCallback(({ groupIdx, startIndex, finishIndex }: { groupIdx: number; startIndex: number; finishIndex: number }) => {
    const newGroups = [...groupsRef.current];
    const group = { ...newGroups[groupIdx] };
    group.links = reorder({
      list: group.links,
      startIndex,
      finishIndex,
    });
    newGroups[groupIdx] = group;
    setGroups(newGroups);
    saveLinks(newGroups);
  }, [saveLinks]);

  const moveLink = useCallback(({
    startGroupIdx,
    finishGroupIdx,
    itemIndexInStartGroup,
    itemIndexInFinishGroup,
  }: {
    startGroupIdx: number;
    finishGroupIdx: number;
    itemIndexInStartGroup: number;
    itemIndexInFinishGroup?: number;
  }) => {
    const newGroups = JSON.parse(JSON.stringify(groupsRef.current));
    const sourceGroup = newGroups[startGroupIdx];
    const destinationGroup = newGroups[finishGroupIdx];
    const [movedLink] = sourceGroup.links.splice(itemIndexInStartGroup, 1);
    
    const destinationIndex = itemIndexInFinishGroup ?? destinationGroup.links.length;
    destinationGroup.links.splice(destinationIndex, 0, movedLink);
    
    setGroups(newGroups);
    saveLinks(newGroups);
  }, [saveLinks]);

  useEffect(() => {
    return combine(
      monitorForElements({
        canMonitor({ source }) {
          return source.data.instanceId === instanceId;
        },
        onDrop({ location, source }) {
          if (!location.current.dropTargets.length) {
            return;
          }

          if (source.data.type === 'group') {
            const startIndex = source.data.groupIdx as number;
            const target = location.current.dropTargets[0];
            const indexOfTarget = target.data.groupIdx as number;
            const closestEdgeOfTarget = extractClosestEdge(target.data);

            const finishIndex = getReorderDestinationIndex({
              startIndex,
              indexOfTarget,
              closestEdgeOfTarget,
              axis: 'vertical',
            });

            reorderGroup({ startIndex, finishIndex });
          }

          if (source.data.type === 'link') {
            const startGroupIdx = source.data.groupIdx as number;
            const startLinkIdx = source.data.linkIdx as number;

            // Dropping on a link
            if (location.current.dropTargets.length === 2) {
              const [linkTarget, groupTarget] = location.current.dropTargets;
              const finishGroupIdx = groupTarget.data.groupIdx as number;
              const finishLinkIdx = linkTarget.data.linkIdx as number;
              const closestEdgeOfTarget = extractClosestEdge(linkTarget.data);

              if (startGroupIdx === finishGroupIdx) {
                      const finishIndex = getReorderDestinationIndex({
                        startIndex: startLinkIdx,
                        indexOfTarget: finishLinkIdx,
                        closestEdgeOfTarget,
                        axis: 'horizontal',
                      });
                      reorderLink({ groupIdx: startGroupIdx, startIndex: startLinkIdx, finishIndex });
                    } else {
                      const finishIndex =
                        closestEdgeOfTarget === 'right' ? finishLinkIdx + 1 : finishLinkIdx;
                      moveLink({
                        startGroupIdx,
                        finishGroupIdx,
                        itemIndexInStartGroup: startLinkIdx,
                        itemIndexInFinishGroup: finishIndex,
                      });
                    }
            }
            // Dropping on a group (empty area or header)
            else if (location.current.dropTargets.length === 1) {
              const [groupTarget] = location.current.dropTargets;
              const finishGroupIdx = groupTarget.data.groupIdx as number;

              if (startGroupIdx !== finishGroupIdx) {
                moveLink({
                  startGroupIdx,
                  finishGroupIdx,
                  itemIndexInStartGroup: startLinkIdx,
                });
              }
            }
          }
        },
      })
    );
  }, [instanceId, reorderGroup, reorderLink, moveLink]);

  const handleAddGroup = () => {
    setCurrentGroup(null);
    setOpen('add-group');
  };

  const handleEditGroup = (group: LinkGroup, index: number) => {
    setCurrentGroup(group);
    setGroupIndex(index);
    setOpen('edit-group');
  };

  const handleDeleteGroup = (index: number) => {
    setGroupIndex(index);
    setOpen('delete-group');
  };

  const handleAddLink = (groupIdx: number) => {
    setGroupIndex(groupIdx);
    setCurrentRow(null);
    setOpen('add');
  };

  const handleEditLink = (groupIdx: number, linkIdx: number, link: Link) => {
    setGroupIndex(groupIdx);
    setLinkIndex(linkIdx);
    setCurrentRow(link);
    setOpen('edit');
  };

  const handleDeleteLink = (groupIdx: number, linkIdx: number) => {
    setGroupIndex(groupIdx);
    setLinkIndex(linkIdx);
    setOpen('delete');
  };

  const onGroupSubmit = (data: LinkGroup) => {
    const newGroups = [...groups];
    if (open === 'add-group') {
      newGroups.push({ ...data, links: [] });
    } else if (open === 'edit-group' && groupIndex !== null) {
      newGroups[groupIndex] = {
        ...newGroups[groupIndex],
        ...data,
      };
    }
    setGroups(newGroups);
    saveLinks(newGroups);
  };

  const onLinkSubmit = (data: Link) => {
    if (groupIndex === null) return;
    const newGroups = [...groups];
    if (open === 'add') {
      newGroups[groupIndex].links.push(data);
    } else if (open === 'edit' && linkIndex !== null) {
      newGroups[groupIndex].links[linkIndex] = data;
    }
    setGroups(newGroups);
    saveLinks(newGroups);
  };

  const onConfirmDelete = () => {
    const newGroups = [...groups];
    if (open === 'delete-group' && groupIndex !== null) {
      newGroups.splice(groupIndex, 1);
    } else if (open === 'delete' && groupIndex !== null && linkIndex !== null) {
      newGroups[groupIndex].links.splice(linkIndex, 1);
    }
    setGroups(newGroups);
    saveLinks(newGroups);
    setOpen(null);
  };

  const boardContextValue = useMemo(() => ({
    getGroups: () => groups,
    reorderGroup,
    reorderLink,
    moveLink,
    instanceId,
    registerGroup: registry.registerGroup,
    registerLink: registry.registerLink,
  }), [groups, reorderGroup, reorderLink, moveLink, instanceId, registry]);

  const handleToggleLinkStatus = useCallback((gIdx: number, lIdx: number) => {
    const newGroups = JSON.parse(JSON.stringify(groupsRef.current));
    const link = newGroups[gIdx].links[lIdx];
    link.status = link.status === 1 ? 0 : 1;
    setGroups(newGroups);
    saveLinks(newGroups);
  }, [saveLinks]);

  return (
    <BoardContext.Provider value={boardContextValue}>
      <ContentLayout
        title="友情链接管理"
        description="管理站点的友情链接及分组。"
        headerActions={
          <Button onClick={handleAddGroup}>
            <Plus className="mr-2 h-4 w-4" /> 添加分组
          </Button>
        }
      >
        {loading ? (
          <div className="flex h-64 items-center justify-center">
            <p className="text-muted-foreground">加载中...</p>
          </div>
        ) : (
          <div className="flex flex-col gap-8 pb-8">
            {groups.map((group, gIdx) => (
              <Column
                key={group.name}
                group={group}
                groupIdx={gIdx}
                onEditGroup={() => handleEditGroup(group, gIdx)}
                onDeleteGroup={() => handleDeleteGroup(gIdx)}
                onAddLink={() => handleAddLink(gIdx)}
                onEditLink={(lIdx) => handleEditLink(gIdx, lIdx, group.links[lIdx])}
                onDeleteLink={(lIdx) => handleDeleteLink(gIdx, lIdx)}
                onToggleLinkStatus={(lIdx) => handleToggleLinkStatus(gIdx, lIdx)}
              />
            ))}
          </div>
        )}

        <LinksGroupActionDialog
          open={open === 'add-group' || open === 'edit-group'}
          onOpenChange={() => setOpen(null)}
          currentRow={currentGroup}
          onSubmit={onGroupSubmit}
        />

        <LinksActionDialog
          open={open === 'add' || open === 'edit'}
          onOpenChange={() => setOpen(null)}
          currentRow={currentRow}
          onSubmit={onLinkSubmit}
        />

        <LinksDeleteDialog
          open={open === 'delete' || open === 'delete-group'}
          onOpenChange={() => setOpen(null)}
          onConfirm={onConfirmDelete}
          title={open === 'delete-group' ? '删除分组' : '删除链接'}
          description={
            open === 'delete-group'
              ? '确定要删除这个分组吗？分组下的所有链接也将被删除。'
              : '确定要删除这个友情链接吗？'
          }
        />
      </ContentLayout>
    </BoardContext.Provider>
  );
}

export default function LinksManagement() {
  return (
    <LinksProvider>
      <LinksManagementContent />
    </LinksProvider>
  );
}
