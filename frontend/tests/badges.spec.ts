import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import StatusBadge from '@/components/tickets/StatusBadge.vue'
import PriorityBadge from '@/components/tickets/PriorityBadge.vue'

describe('StatusBadge', () => {
  it.each([
    ['open', 'Open'],
    ['in_progress', 'In Progress'],
    ['resolved', 'Resolved'],
    ['closed', 'Closed'],
  ] as const)('renders label for %s', (status, label) => {
    const wrapper = mount(StatusBadge, { props: { status } })
    expect(wrapper.text()).toContain(label)
  })
})

describe('PriorityBadge', () => {
  it.each([
    ['low', 'Low'],
    ['medium', 'Medium'],
    ['high', 'High'],
    ['urgent', 'Urgent'],
  ] as const)('renders label for %s', (priority, label) => {
    const wrapper = mount(PriorityBadge, { props: { priority } })
    expect(wrapper.text()).toContain(label)
  })
})
