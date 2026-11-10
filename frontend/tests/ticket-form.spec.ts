import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import TicketForm from '@/components/tickets/TicketForm.vue'
import type { TicketFormErrors as FormErrors } from '@/lib/types'
import type { CreateTicketPayload } from '@/stores/tickets'

async function typeTitle(wrapper: ReturnType<typeof mount>, value: string) {
  await wrapper.find('input').setValue(value)
}

async function typeDescription(wrapper: ReturnType<typeof mount>, value: string) {
  await wrapper.find('textarea').setValue(value)
}

describe('TicketForm validation', () => {
  it('shows a dedicated "wajib diisi" message for empty required fields and emits nothing', async () => {
    const wrapper = mount(TicketForm)
    await wrapper.find('form').trigger('submit.prevent')
    expect(wrapper.find('input + p, [id*="-error"]').exists()).toBe(true)
    expect(wrapper.text()).toContain('Judul wajib diisi.')
    expect(wrapper.text()).toContain('Deskripsi wajib diisi.')
    expect(wrapper.emitted('submit')).toBeUndefined()
  })

  it('marks every required field with a * and a "wajib diisi" note', () => {
    const wrapper = mount(TicketForm)
    // Judul, Deskripsi, Kategori, Prioritas — legend markers excluded.
    const labelStars = wrapper.findAll('label span.text-red-600[aria-hidden="true"]')
    expect(labelStars.length).toBe(4)
    expect(wrapper.text()).toContain('wajib diisi.')
  })

  it('rejects a title shorter than 3 characters (non-empty message)', async () => {
    const wrapper = mount(TicketForm)
    await typeTitle(wrapper, 'ab')
    await typeDescription(wrapper, 'Deskripsi memadai untuk tiket baru.')
    await wrapper.find('form').trigger('submit.prevent')
    expect(wrapper.text()).toContain('Judul minimal 3 karakter.')
    expect(wrapper.emitted('submit')).toBeUndefined()
  })

  it('rejects a description with fewer than 10 characters', async () => {
    const wrapper = mount(TicketForm)
    await typeTitle(wrapper, 'Judul yang layak')
    await typeDescription(wrapper, 'terlalu')
    await wrapper.find('form').trigger('submit.prevent')
    expect(wrapper.text()).toContain('Deskripsi minimal 10 karakter.')
    expect(wrapper.emitted('submit')).toBeUndefined()
  })

  it('emits a normalized payload (trimmed) once everything is valid', async () => {
    const wrapper = mount(TicketForm)
    await typeTitle(wrapper, '  Laptop lab gagal boot  ')
    await typeDescription(wrapper, ' Layar hitam setelah memilih opsi boot. ')
    await wrapper.find('form').trigger('submit.prevent')
    const emitted = wrapper.emitted('submit') as unknown as Array<[CreateTicketPayload]>
    expect(emitted).toHaveLength(1)
    expect(emitted[0][0]).toEqual({
      title: 'Laptop lab gagal boot',
      description: 'Layar hitam setelah memilih opsi boot.',
      category: 'other',
      priority: 'medium',
    })
  })

  it('renders server-side validation errors passed through the errors prop', () => {
    const errors: FormErrors = { title: 'Judul sudah dipakai.' }
    const wrapper = mount(TicketForm, { props: { errors } })
    expect(wrapper.text()).toContain('Judul sudah dipakai.')
  })

  it('disables both buttons while loading to prevent double submit', () => {
    const wrapper = mount(TicketForm, { props: { loading: true } })
    const buttons = wrapper.findAll('button')
    expect(buttons.every((b) => b.attributes('disabled') !== undefined)).toBe(true)
  })
})
