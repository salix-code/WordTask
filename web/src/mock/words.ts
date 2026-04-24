import type { Word } from '@/types/word'

/**
 * M1 阶段使用的 Mock 单词数据（M2 会被 fetchTodayWords 替换）
 */
export const mockWords: Word[] = [
  {
    id: 'w-001',
    text: 'ephemeral',
    phonetic: '/ɪˈfem.ɚ.əl/',
    definitions: [
      { pos: 'adj.', meaning: '短暂的；瞬息的' },
    ],
    examples: [
      { en: 'Fame in the internet age can be ephemeral.', zh: '网络时代的名声可能是短暂的。' },
    ],
  },
  {
    id: 'w-002',
    text: 'resilient',
    phonetic: '/rɪˈzɪl.i.ənt/',
    definitions: [
      { pos: 'adj.', meaning: '有复原力的；适应力强的' },
    ],
    examples: [
      { en: 'Children are often remarkably resilient.', zh: '孩子们往往有惊人的适应力。' },
    ],
  },
  {
    id: 'w-003',
    text: 'meticulous',
    phonetic: '/məˈtɪk.jə.ləs/',
    definitions: [{ pos: 'adj.', meaning: '一丝不苟的；小心翼翼的' }],
    examples: [
      { en: 'He keeps meticulous notes.', zh: '他的笔记非常仔细。' },
    ],
  },
  {
    id: 'w-004',
    text: 'pragmatic',
    phonetic: '/præɡˈmæt.ɪk/',
    definitions: [{ pos: 'adj.', meaning: '务实的；实用主义的' }],
    examples: [{ en: 'We need a pragmatic approach.', zh: '我们需要一种务实的方法。' }],
  },
  {
    id: 'w-005',
    text: 'serendipity',
    phonetic: '/ˌser.ənˈdɪp.ə.ti/',
    definitions: [{ pos: 'n.', meaning: '意外发现美好事物的能力' }],
    examples: [{ en: 'A lucky case of serendipity.', zh: '一次幸运的意外收获。' }],
  },
  {
    id: 'w-006',
    text: 'ubiquitous',
    phonetic: '/juːˈbɪk.wə.təs/',
    definitions: [{ pos: 'adj.', meaning: '无处不在的' }],
    examples: [{ en: 'Mobile phones are now ubiquitous.', zh: '手机现在无处不在。' }],
  },
  {
    id: 'w-007',
    text: 'candid',
    phonetic: '/ˈkæn.dɪd/',
    definitions: [{ pos: 'adj.', meaning: '坦率的；直言不讳的' }],
    examples: [{ en: 'To be candid, I disagree.', zh: '坦率地说，我不同意。' }],
  },
]
