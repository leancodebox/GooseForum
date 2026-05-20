import { useEffect, useState, useCallback, useMemo } from 'react';
import { toast } from 'sonner';
import { Save, Loader2, PenLine } from 'lucide-react';
import { combine } from '@atlaskit/pragmatic-drag-and-drop/combine';
import { monitorForElements } from '@atlaskit/pragmatic-drag-and-drop/element/adapter';
import { reorder } from '@atlaskit/pragmatic-drag-and-drop/reorder';
import { extractClosestEdge } from '@atlaskit/pragmatic-drag-and-drop-hitbox/closest-edge';
import { getReorderDestinationIndex } from '@atlaskit/pragmatic-drag-and-drop-hitbox/util/get-reorder-destination-index';

import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { ContentLayout } from '@/components/layout/content-layout';
import SponsorshipProvider, { useSponsorship } from './components/sponsorship-provider';
import { SponsorActionDialog } from './components/sponsor-action-dialog';
import { type SponsorsConfig, type SponsorItem, type Sponsors } from './data/schema';
import { SponsorColumn } from './components/sponsor-column';
import { BoardContext } from './components/board-context';
import { createRegistry } from './components/registry';
import { getSponsors, saveSponsors } from '@/api';

const LEVEL_NAMES: Record<string, string> = {
  level0: '特别赞助商 (Level 0)',
  level1: '金牌赞助商 (Level 1)',
  level2: '银牌赞助商 (Level 2)',
  level3: '铜牌赞助商 (Level 3)',
}

const defaultSponsorsConfig: SponsorsConfig = {
  sponsors: {
    level0: [],
    level1: [],
    level2: [],
    level3: [],
  },
  content: {
    title: '赞助',
    description: '感谢这些赞助者帮助 GooseForum 持续变好。',
  },
  contact: {
    title: '成为赞助者',
    description: '支持社区建设，赞助者可展示在赞助页，并获得更醒目的社区露出。',
    buttonText: '联系我们',
    buttonLink: 'mailto:contact@gooseforum.online',
  },
  rules: [
    { content: '链接需稳定可访问。' },
    { content: '内容需适合公开社区展示。' },
    { content: '头像或 Logo 建议保持清晰。' },
  ],
}

function normalizeSponsorsConfig(config?: Partial<SponsorsConfig>): SponsorsConfig {
  return {
    sponsors: {
      level0: config?.sponsors?.level0 ?? [],
      level1: config?.sponsors?.level1 ?? [],
      level2: config?.sponsors?.level2 ?? [],
      level3: config?.sponsors?.level3 ?? [],
    },
    content: {
      title: config?.content?.title || defaultSponsorsConfig.content.title,
      description: config?.content?.description || defaultSponsorsConfig.content.description,
    },
    contact: {
      title: config?.contact?.title || defaultSponsorsConfig.contact.title,
      description: config?.contact?.description || defaultSponsorsConfig.contact.description,
      buttonText: config?.contact?.buttonText || defaultSponsorsConfig.contact.buttonText,
      buttonLink: config?.contact?.buttonLink || defaultSponsorsConfig.contact.buttonLink,
    },
    rules: config?.rules ?? defaultSponsorsConfig.rules,
  }
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

  const [config, setConfig] = useState<SponsorsConfig>(defaultSponsorsConfig);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [instanceId] = useState(() => Symbol('instance-id'));
  const [registry] = useState(createRegistry);

  const fetchSponsors = async () => {
    try {
      const res = await getSponsors();
      if (res.code === 0) {
        setConfig(normalizeSponsorsConfig(res.result));
      } else {
        toast.error(res.msg || '获取赞助商失败');
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
      const res = await saveSponsors({
        sponsorsInfo: normalizeSponsorsConfig(config),
      });
      if (res.code === 0) {
        toast.success('保存成功');
      } else {
        toast.error(res.msg || '保存失败');
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

  const updateContent = (key: keyof SponsorsConfig['content'], value: string) => {
    setConfig((prev) => ({
      ...prev,
      content: {
        ...prev.content,
        [key]: value,
      },
    }));
  };

  const updateContact = (key: keyof SponsorsConfig['contact'], value: string) => {
    setConfig((prev) => ({
      ...prev,
      contact: {
        ...prev.contact,
        [key]: value,
      },
    }));
  };

  const updateRule = (index: number, value: string) => {
    setConfig((prev) => ({
      ...prev,
      rules: prev.rules.map((rule, ruleIndex) => (
        ruleIndex === index ? { content: value } : rule
      )),
    }));
  };

  const addRule = () => {
    setConfig((prev) => ({
      ...prev,
      rules: [...prev.rules, { content: '' }],
    }));
  };

  const removeRule = (index: number) => {
    setConfig((prev) => ({
      ...prev,
      rules: prev.rules.filter((_, ruleIndex) => ruleIndex !== index),
    }));
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
              <section className='border-b border-border/70 pb-4'>
                <div className='flex flex-wrap items-center gap-2'>
                  <div className='group relative min-w-[180px] flex-1 rounded-md transition-colors hover:bg-muted/50 focus-within:bg-muted/50'>
                    <Input
                      aria-label='赞助页标题'
                      title='点击编辑赞助页标题'
                      className='h-auto border-0 bg-transparent px-2 py-1 pr-9 text-2xl font-bold tracking-tight shadow-none focus-visible:ring-1'
                      value={config.content.title}
                      onChange={(event) => updateContent('title', event.target.value)}
                    />
                    <PenLine className='pointer-events-none absolute right-2 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground/40 opacity-0 transition-opacity group-hover:opacity-100 group-focus-within:opacity-100' />
                  </div>
                  <span className='rounded-full bg-muted px-2 py-0.5 text-xs font-semibold text-muted-foreground'>
                    {Object.values(config.sponsors).reduce((total, sponsors) => total + sponsors.length, 0)}
                  </span>
                </div>
                <div className='group relative mt-2 rounded-md transition-colors hover:bg-muted/50 focus-within:bg-muted/50'>
                  <Input
                    aria-label='赞助页描述'
                    title='点击编辑赞助页描述'
                    className='h-10 border-0 bg-transparent px-2 pr-9 text-sm text-muted-foreground shadow-none focus-visible:ring-1'
                    value={config.content.description}
                    onChange={(event) => updateContent('description', event.target.value)}
                  />
                  <PenLine className='pointer-events-none absolute right-2 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground/40 opacity-0 transition-opacity group-hover:opacity-100 group-focus-within:opacity-100' />
                </div>
              </section>
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
                <h2 className='text-sm font-semibold text-foreground'>联系入口</h2>
                <div className='mt-4 space-y-3'>
                  <div className='group relative rounded-md transition-colors hover:bg-muted/50 focus-within:bg-muted/50'>
                    <Input
                      aria-label='联系入口标题'
                      title='点击编辑联系入口标题'
                      className='h-9 border-0 bg-transparent px-2 pr-8 text-sm font-semibold shadow-none focus-visible:ring-1'
                      value={config.contact.title}
                      onChange={(event) => updateContact('title', event.target.value)}
                    />
                    <PenLine className='pointer-events-none absolute right-2 top-1/2 h-3.5 w-3.5 -translate-y-1/2 text-muted-foreground/40 opacity-0 transition-opacity group-hover:opacity-100 group-focus-within:opacity-100' />
                  </div>
                  <div className='group relative rounded-md transition-colors hover:bg-muted/50 focus-within:bg-muted/50'>
                    <Textarea
                      aria-label='联系入口描述'
                      title='点击编辑联系入口描述'
                      className='min-h-20 resize-none border-0 bg-transparent px-2 py-2 pr-8 text-sm leading-6 text-muted-foreground shadow-none focus-visible:ring-1'
                      value={config.contact.description}
                      onChange={(event) => updateContact('description', event.target.value)}
                    />
                    <PenLine className='pointer-events-none absolute right-2 top-2.5 h-3.5 w-3.5 text-muted-foreground/40 opacity-0 transition-opacity group-hover:opacity-100 group-focus-within:opacity-100' />
                  </div>
                  <div className='grid gap-2 sm:grid-cols-2 xl:grid-cols-1'>
                    <div className='space-y-1'>
                      <div className='px-2 text-[11px] font-semibold uppercase text-muted-foreground'>按钮文字</div>
                      <div className='group relative rounded-md transition-colors hover:bg-muted/50 focus-within:bg-muted/50'>
                        <Input
                          aria-label='联系入口按钮文字'
                          title='点击编辑按钮文字'
                          className='h-9 border-0 bg-transparent px-2 pr-8 text-sm shadow-none focus-visible:ring-1'
                          value={config.contact.buttonText}
                          onChange={(event) => updateContact('buttonText', event.target.value)}
                        />
                        <PenLine className='pointer-events-none absolute right-2 top-1/2 h-3.5 w-3.5 -translate-y-1/2 text-muted-foreground/40 opacity-0 transition-opacity group-hover:opacity-100 group-focus-within:opacity-100' />
                      </div>
                    </div>
                    <div className='space-y-1'>
                      <div className='px-2 text-[11px] font-semibold uppercase text-muted-foreground'>按钮链接</div>
                      <div className='group relative rounded-md transition-colors hover:bg-muted/50 focus-within:bg-muted/50'>
                        <Input
                          aria-label='联系入口按钮链接'
                          title='点击编辑按钮链接'
                          className='h-9 border-0 bg-transparent px-2 pr-8 text-sm shadow-none focus-visible:ring-1'
                          placeholder='mailto:contact@example.com'
                          value={config.contact.buttonLink}
                          onChange={(event) => updateContact('buttonLink', event.target.value)}
                        />
                        <PenLine className='pointer-events-none absolute right-2 top-1/2 h-3.5 w-3.5 -translate-y-1/2 text-muted-foreground/40 opacity-0 transition-opacity group-hover:opacity-100 group-focus-within:opacity-100' />
                      </div>
                    </div>
                  </div>
                </div>
              </div>
              <div className='rounded-lg border border-border/70 bg-background p-4'>
                <div className='flex items-center justify-between gap-2'>
                  <h2 className='text-sm font-semibold text-foreground'>展示规则</h2>
                  <Button type='button' size='sm' variant='outline' onClick={addRule}>
                    添加
                  </Button>
                </div>
                <div className='mt-4 space-y-2'>
                  {config.rules.length === 0 ? (
                    <p className='text-sm text-muted-foreground'>暂无展示规则，前台会隐藏该区块。</p>
                  ) : (
                    config.rules.map((rule, index) => (
                      <div key={index} className='flex gap-2'>
                        <div className='group relative min-w-0 flex-1 rounded-md transition-colors hover:bg-muted/50 focus-within:bg-muted/50'>
                          <Input
                            aria-label={`展示规则 ${index + 1}`}
                            title='点击编辑展示规则'
                            className='h-9 border-0 bg-transparent px-2 pr-8 text-sm shadow-none focus-visible:ring-1'
                            value={rule.content}
                            onChange={(event) => updateRule(index, event.target.value)}
                          />
                          <PenLine className='pointer-events-none absolute right-2 top-1/2 h-3.5 w-3.5 -translate-y-1/2 text-muted-foreground/40 opacity-0 transition-opacity group-hover:opacity-100 group-focus-within:opacity-100' />
                        </div>
                        <Button
                          type='button'
                          variant='ghost'
                          size='sm'
                          className='text-destructive hover:text-destructive'
                          onClick={() => removeRule(index)}
                        >
                          删除
                        </Button>
                      </div>
                    ))
                  )}
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
