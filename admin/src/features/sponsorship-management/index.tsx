import { useEffect, useState, useCallback, useMemo } from 'react';
import axios from 'axios';
import { toast } from 'sonner';
import { Save, Loader2 } from 'lucide-react';
import { combine } from '@atlaskit/pragmatic-drag-and-drop/combine';
import { monitorForElements } from '@atlaskit/pragmatic-drag-and-drop/element/adapter';
import { reorder } from '@atlaskit/pragmatic-drag-and-drop/reorder';
import { extractClosestEdge } from '@atlaskit/pragmatic-drag-and-drop-hitbox/closest-edge';
import { getReorderDestinationIndex } from '@atlaskit/pragmatic-drag-and-drop-hitbox/util/get-reorder-destination-index';

import { Button } from '@/components/ui/button';
import { ContentLayout } from '@/components/layout/content-layout';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import SponsorshipProvider, { useSponsorship } from './components/sponsorship-provider';
import { SponsorActionDialog } from './components/sponsor-action-dialog';
import { UserSponsorActionDialog } from './components/user-sponsor-action-dialog';
import { type SponsorsConfig, type SponsorItem, type UserSponsor, type Sponsors } from './data/schema';
import { SponsorColumn } from './components/sponsor-column';
import { UserColumn } from './components/user-column';
import { BoardContext } from './components/board-context';
import { createRegistry } from './components/registry';

const LEVEL_NAMES: Record<string, string> = {
  level0: '特别赞助商 (Level 0)',
  level1: '金牌赞助商 (Level 1)',
  level2: '银牌赞助商 (Level 2)',
  level3: '铜牌赞助商 (Level 3)',
}

function SponsorshipManagementContent() {
  const {
    open,
    setOpen,
    currentRow,
    setCurrentRow,
    currentLevel,
    setCurrentLevel,
    currentIndex,
    setCurrentIndex,
  } = useSponsorship();

  const [config, setConfig] = useState<SponsorsConfig>({
    sponsors: {
      level0: [],
      level1: [],
      level2: [],
      level3: [],
    },
    users: [],
  });
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [instanceId] = useState(() => Symbol('instance-id'));
  const [registry] = useState(createRegistry);

  const fetchSponsors = async () => {
    try {
      const res = await axios.get('/api/admin/sponsors');
      if (res.data.code === 0) {
        setConfig(res.data.result || { sponsors: { level0: [], level1: [], level2: [], level3: [] }, users: [] });
      } else {
        toast.error(res.data.msg || '获取赞助商失败');
      }
    } catch (error) {
      toast.error('获取赞助商失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchSponsors();
  }, []);

  const handleSave = async () => {
    setSaving(true);
    try {
      const res = await axios.post('/api/admin/save-sponsors', {
        sponsorsInfo: config,
      });
      if (res.data.code === 0) {
        toast.success('保存成功');
      } else {
        toast.error(res.data.msg || '保存失败');
      }
    } catch (error) {
      toast.error('保存失败');
    } finally {
      setSaving(false);
    }
  };

  const reorderSponsor = useCallback(({ sourceLevel, destLevel, startIndex, finishIndex }: { sourceLevel: keyof Sponsors; destLevel: keyof Sponsors; startIndex: number; finishIndex: number }) => {
    setConfig((prev) => {
      const newSponsors = {
        level0: [...prev.sponsors.level0],
        level1: [...prev.sponsors.level1],
        level2: [...prev.sponsors.level2],
        level3: [...prev.sponsors.level3],
      };
      
      if (sourceLevel === destLevel) {
        newSponsors[sourceLevel] = reorder({
          list: newSponsors[sourceLevel],
          startIndex,
          finishIndex,
        });
      } else {
        const [movedItem] = newSponsors[sourceLevel].splice(startIndex, 1);
        newSponsors[destLevel].splice(finishIndex, 0, movedItem);
      }
      
      return { ...prev, sponsors: newSponsors };
    });
  }, []);

  const reorderUser = useCallback(({ startIndex, finishIndex }: { startIndex: number; finishIndex: number }) => {
    setConfig((prev) => {
      const newUsers = reorder({
        list: prev.users,
        startIndex,
        finishIndex,
      });
      return { ...prev, users: newUsers };
    });
  }, []);

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

          if (source.data.type === 'sponsor') {
            const startLevel = source.data.level as keyof Sponsors;
            const startIndex = source.data.index as number;

            // Dropping on a sponsor card
            if (location.current.dropTargets.length >= 2) {
              const sponsorTarget = location.current.dropTargets.find(t => t.data.type === 'sponsor');
              const levelTarget = location.current.dropTargets.find(t => t.data.type === 'level');
              
              if (sponsorTarget && levelTarget) {
                const destLevel = levelTarget.data.level as keyof Sponsors;
                const destIndex = sponsorTarget.data.index as number;
                const closestEdgeOfTarget = extractClosestEdge(sponsorTarget.data);

                const finishIndex = getReorderDestinationIndex({
                  startIndex: startLevel === destLevel ? startIndex : -1,
                  indexOfTarget: destIndex,
                  closestEdgeOfTarget,
                  axis: 'horizontal',
                });

                reorderSponsor({ sourceLevel: startLevel, destLevel, startIndex, finishIndex });
              }
            }
            // Dropping on a level column (empty area or header)
            else if (location.current.dropTargets.length === 1) {
              const [levelTarget] = location.current.dropTargets;
              const destLevel = levelTarget.data.level as keyof Sponsors;

              if (startLevel !== destLevel) {
                reorderSponsor({
                  sourceLevel: startLevel,
                  destLevel,
                  startIndex,
                  finishIndex: config.sponsors[destLevel].length,
                });
              }
            }
          }

          if (source.data.type === 'user') {
            const startIndex = source.data.index as number;

            // Dropping on a user card
            if (location.current.dropTargets.length >= 1) {
              const target = location.current.dropTargets[0];
              
              if (target.data.type === 'user') {
                const destIndex = target.data.index as number;
                const closestEdgeOfTarget = extractClosestEdge(target.data);

                const finishIndex = getReorderDestinationIndex({
                  startIndex,
                  indexOfTarget: destIndex,
                  closestEdgeOfTarget,
                  axis: 'horizontal',
                });

                reorderUser({ startIndex, finishIndex });
              }
            }
          }
        },
      })
    );
  }, [instanceId, reorderSponsor, reorderUser, config.sponsors]);

  const contextValue = useMemo(() => ({
    getConfig: () => config,
    reorderSponsor,
    reorderUser,
    instanceId,
    registerLevel: registry.registerLevel,
    registerSponsor: registry.registerSponsor,
    registerUser: registry.registerUser,
  }), [config, reorderSponsor, reorderUser, instanceId, registry]);

  const handleAddSponsor = (level: keyof Sponsors) => {
    setCurrentLevel(level);
    setCurrentIndex(null);
    setCurrentRow(null);
    setOpen('add-sponsor');
  };

  const handleEditSponsor = (level: keyof Sponsors, index: number) => {
    setCurrentLevel(level);
    setCurrentIndex(index);
    setCurrentRow(config.sponsors[level][index]);
    setOpen('edit-sponsor');
  };

  const handleDeleteSponsor = (level: keyof Sponsors, index: number) => {
    setConfig((prev) => {
      const newSponsors = { ...prev.sponsors };
      newSponsors[level].splice(index, 1);
      return { ...prev, sponsors: newSponsors };
    });
  };

  const handleSponsorSubmit = (data: SponsorItem) => {
    setConfig((prev) => {
      const newSponsors = { ...prev.sponsors };
      if (currentLevel) {
        if (currentIndex !== null) {
          newSponsors[currentLevel][currentIndex] = data;
        } else {
          newSponsors[currentLevel].push(data);
        }
      }
      return { ...prev, sponsors: newSponsors };
    });
    setOpen(null);
  };

  const handleAddUser = () => {
    setCurrentIndex(null);
    setCurrentRow(null);
    setOpen('add-user');
  };

  const handleEditUser = (index: number) => {
    setCurrentIndex(index);
    setCurrentRow(config.users[index]);
    setOpen('edit-user');
  };

  const handleDeleteUser = (index: number) => {
    setConfig((prev) => {
      const newUsers = [...prev.users];
      newUsers.splice(index, 1);
      return { ...prev, users: newUsers };
    });
  };

  const handleUserSubmit = (data: UserSponsor) => {
    setConfig((prev) => {
      const newUsers = [...prev.users];
      if (currentIndex !== null) {
        newUsers[currentIndex] = data;
      } else {
        newUsers.push(data);
      }
      return { ...prev, users: newUsers };
    });
    setOpen(null);
  };

  return (
    <BoardContext.Provider value={contextValue}>
      <ContentLayout
        title='赞助管理'
        description='管理赞助商和个人赞助者。'
        headerActions={
          <Button onClick={handleSave} disabled={saving}>
            {saving ? (
              <Loader2 className='mr-2 h-4 w-4 animate-spin' />
            ) : (
              <Save className='mr-2 h-4 w-4' />
            )}
            保存配置
          </Button>
        }
      >
        {loading ? (
          <div className='flex h-64 items-center justify-center'>
            <Loader2 className='h-8 w-8 animate-spin text-primary' />
          </div>
        ) : (
          <Tabs defaultValue='sponsors' className='w-full'>
            <TabsList className='mb-4'>
              <TabsTrigger value='sponsors'>赞助商管理</TabsTrigger>
              <TabsTrigger value='users'>个人赞助者管理</TabsTrigger>
            </TabsList>

            <TabsContent value='sponsors' className='space-y-6'>
              <div className='flex flex-col gap-6'>
                {(Object.keys(LEVEL_NAMES) as Array<keyof Sponsors>).map((level) => (
                  <SponsorColumn
                    key={level}
                    level={level}
                    title={LEVEL_NAMES[level]}
                    sponsors={config.sponsors[level]}
                    onAdd={() => handleAddSponsor(level)}
                    onEdit={(index) => handleEditSponsor(level, index)}
                    onDelete={(index) => handleDeleteSponsor(level, index)}
                  />
                ))}
              </div>
            </TabsContent>

            <TabsContent value='users'>
              <UserColumn
                users={config.users}
                onAdd={handleAddUser}
                onEdit={handleEditUser}
                onDelete={handleDeleteUser}
              />
            </TabsContent>
          </Tabs>
        )}

        <SponsorActionDialog
          open={open === 'add-sponsor' || open === 'edit-sponsor'}
          onOpenChange={() => setOpen(null)}
          onSubmit={handleSponsorSubmit}
          currentRow={currentRow as SponsorItem}
          title={open === 'add-sponsor' ? '添加赞助商' : '编辑赞助商'}
        />

        <UserSponsorActionDialog
          open={open === 'add-user' || open === 'edit-user'}
          onOpenChange={() => setOpen(null)}
          onSubmit={handleUserSubmit}
          currentRow={currentRow as UserSponsor}
          title={open === 'add-user' ? '添加个人赞助' : '编辑个人赞助'}
        />
      </ContentLayout>
    </BoardContext.Provider>
  );
}

export default function SponsorshipManagement() {
  return (
    <SponsorshipProvider>
      <SponsorshipManagementContent />
    </SponsorshipProvider>
  );
}
