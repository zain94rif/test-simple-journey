import { beforeEach, describe, expect, it } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { mount } from '@vue/test-utils'
import RoleSelect from '@/components/users/RoleSelect.vue'
import { useAuthStore } from '@/stores/auth'
import type { User } from '@/lib/types'

vi.mock('@/lib/api', () => ({
  getSingle: vi.fn(),
  getList: vi.fn(),
  postSingle: vi.fn(),
  patchSingle: vi.fn(),
}))

function makeUser(id: number, role = 'agent'): User {
  return { id, name: `User ${id}`, email: `u${id}@x`, role, created_at: '' } as User
}

describe('RoleSelect', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('is enabled when the selected user is not the current user', () => {
    const auth = useAuthStore()
    auth.user = makeUser(1, 'admin')
    const wrapper = mount(RoleSelect, { props: { user: makeUser(2) } })
    const select = wrapper.find('select')
    expect(select.attributes('disabled')).toBeUndefined()
  })

  it('is disabled for the current user (self-role rule) and shows a hint', () => {
    const auth = useAuthStore()
    auth.user = makeUser(1, 'admin')
    const wrapper = mount(RoleSelect, { props: { user: makeUser(1, 'admin') } })
    const select = wrapper.find('select')
    expect(select.attributes('disabled')).toBeDefined()
    expect(wrapper.find('[title]').attributes('title')).toContain('tidak dapat mengubah peran sendiri')
    // No change event emitted when disabled.
    select.setValue('agent')
    expect(wrapper.emitted('change')).toBeUndefined()
  })

  it('emits the chosen role for another user', async () => {
    const auth = useAuthStore()
    auth.user = makeUser(1, 'admin')
    const wrapper = mount(RoleSelect, { props: { user: makeUser(4) } })
    const select = wrapper.find('select')
    await select.setValue('requester')
    expect(wrapper.emitted('change')).toBeTruthy()
    expect(wrapper.emitted('change')!.at(-1)).toEqual(['requester'])
    expect(wrapper.emitted('change')!.length).toBe(1)
    select.setValue('admin')
    expect(wrapper.emitted('change')!.at(-1)).toEqual(['admin'])
  })
})

export {}
