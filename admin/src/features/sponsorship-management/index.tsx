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
import SponsorshipProvider, { useSponsorship } from './components/sponsorship-provider';
import { SponsorActionDialog } from './components/sponsor-action-dialog';
import { type SponsorsConfig, type SponsorItem, type Sponsors } from './data/schema';
import { SponsorColumn } from './components/sponsor-column';
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
  });
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [instanceId] = useState(() => Symbol('instance-id'));
  const [registry] = useState(createRegistry);

  const fetchSponsors = async () => {
    try {
      const res = await axios.get('/api/admin/sponsors');
      if (res.data.code === 0) {
        setConfig(res.data.result || { sponsors: { level0: [], level1: [], level2: [], level3: [] } });
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

        },
      })
    );
  }, [instanceId, reorderSponsor, config.sponsors]);

  const contextValue = useMemo(() => ({
    getConfig: () => config,
    reorderSponsor,
    instanceId,
    registerLevel: registry.registerLevel,
    registerSponsor: registry.registerSponsor,
  }), [config, reorderSponsor, instanceId, registry]);

  const handleAddSponsor = (level: keyof Sponsors) => {
    setCurrentLevel(level);
    setCurrentIndex(null);
    setCurrentRow({
      name: '',
      avatarUrl: '',
      message: '',
      link: ''
    });
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

  return (
    <BoardContext.Provider value={contextValue}>
      <ContentLayout
        title='赞助管理'
        description='按前台赞助页的展示层级管理赞助商。'
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
          <div className='grid gap-4 xl:grid-cols-[minmax(0,1fr)_260px]'>
            <div className='space-y-5'>
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
            <aside className='space-y-3'>
              <div className='rounded-lg border border-border/70 bg-background p-4'>
                <h2 className='text-sm font-semibold text-foreground'>展示规则</h2>
                <div className='mt-3 space-y-2 text-sm leading-6 text-muted-foreground'>
                  <p>层级本身固定，不参与拖拽。</p>
                  <p>同一层级内的赞助商可以拖拽排序，也可以拖到其他层级。</p>
                  <p>保存后会同步到前台赞助页。</p>
                </div>
              </div>
              <div className='rounded-lg border border-border/70 bg-background p-4'>
                <h2 className='text-sm font-semibold text-foreground'>内容建议</h2>
                <div className='mt-3 space-y-2 text-sm leading-6 text-muted-foreground'>
                  <p>Logo 建议使用清晰方图。</p>
                  <p>描述文案保持一到两行，前台会做截断。</p>
                  <p>链接留空时前台不会跳转。</p>
                </div>
              </div>
            </aside>
          </div>
        )}

        <SponsorActionDialog
          open={open === 'add-sponsor' || open === 'edit-sponsor'}
          onOpenChange={() => setOpen(null)}
          onSubmit={handleSponsorSubmit}
          currentRow={currentRow as SponsorItem}
          title={open === 'add-sponsor' ? '添加赞助商' : '编辑赞助商'}
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
