import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import SkeletonBlock from '@/components/ui/SkeletonBlock.vue'

/**
 * Regression: `withDefaults(defineProps ...)` tidak pernah di-assign ke variabel, sehingga
 * template merujuk `props` yang tidak pernah ada → render crash (UI beku walau data
 * data sudah diterima dari jaringan).
 */
describe('SkeletonBlock', () => {
  it('renders without crashing when data is loading', () => {
    expect(() => mount(SkeletonBlock)).not.toThrow()
  })

  it('renders with width and height styles', () => {
    const wrapper = mount(SkeletonBlock, { props: { width: '50%', height: '24px' } })
    const el = wrapper.get('div')
    expect(el.attributes('style')).toContain('width: 50%')
    expect(el.attributes('style')).toContain('height: 24px')
    expect(el.classes()).toContain('animate-pulse')
  })
})
