import { beforeEach, describe, expect, it, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { mount } from '@vue/test-utils'
import CommentForm from '@/components/tickets/CommentForm.vue'
import { useAuthStore } from '@/stores/auth'
import { useTicketStore } from '@/stores/tickets'
import type { Ticket, User } from '@/lib/types'

vi.mock('@/lib/api', () => ({
  getSingle: vi.fn(),
  getList: vi.fn(),
  postSingle: vi.fn(),
  patchSingle: vi.fn(),
}))

function stubTicket(): Ticket {
  return { id: 7, status: 'in_progress', comments: [] } as unknown as Ticket
}

async function fillAndSubmit(wrapper: Awaited<ReturnType<typeof mount>>) {
  await wrapper.find('textarea').setValue('Mohon verifikasi kembali.')
  await wrapper.find('form').trigger('submit.prevent')
}

describe('CommentForm', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('blocks empty comments with a validation error', async () => {
    const wrapper = mount(CommentForm, { props: { ticket: stubTicket() } })
    await wrapper.find('form').trigger('submit.prevent')
    expect(wrapper.text()).toContain('Komentar tidak boleh kosong.')
  })

  it('requester has no is-internal checkbox and submits a public comment', async () => {
    const auth = useAuthStore()
    auth.user = { id: 10, name: 'R', email: 'r@x', role: 'requester', created_at: '' } as User

    const store = useTicketStore()
    const addComment = vi.fn().mockResolvedValue({})
    store.addComment = addComment

    const wrapper = mount(CommentForm, { props: { ticket: stubTicket() } })
    expect(wrapper.find('input[type="checkbox"]').exists()).toBe(false)

    await fillAndSubmit(wrapper)
    expect(addComment).toHaveBeenCalledTimes(1)
    expect(addComment).toHaveBeenCalledWith(7, 'Mohon verifikasi kembali.', false)
  })

  it('staff can toggle is-internal and it is passed to the action', async () => {
    const auth = useAuthStore()
    auth.user = { id: 2, name: 'A', email: 'a@x', role: 'agent', created_at: '' } as User

    const store = useTicketStore()
    const addComment = vi.fn().mockResolvedValue({})
    store.addComment = addComment

    const wrapper = mount(CommentForm, { props: { ticket: stubTicket() } })
    const checkbox = wrapper.find('input[type="checkbox"]')
    expect(checkbox.exists()).toBe(true)
    await checkbox.setValue(true)
    await fillAndSubmit(wrapper)
    expect(addComment).toHaveBeenCalledWith(7, 'Mohon verifikasi kembali.', true)
  })

  it('clears the textarea after a successful send', async () => {
    const auth = useAuthStore()
    auth.user = { id: 10, name: 'R', email: 'r@x', role: 'requester', created_at: '' } as User
    const store = useTicketStore()
    store.addComment = vi.fn().mockResolvedValue({})

    const wrapper = mount(CommentForm, { props: { ticket: stubTicket() } })
    await fillAndSubmit(wrapper)
    expect((wrapper.find('textarea').element as HTMLTextAreaElement).value).toBe('')
  })
})
