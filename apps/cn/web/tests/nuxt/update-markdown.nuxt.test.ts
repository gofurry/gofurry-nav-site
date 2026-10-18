import { it, expect } from 'vitest'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import UpdateMarkdown from '../../app/components/updates/UpdateMarkdown.vue'

it('renders sanitized Markdown through the real Vue boundary and reacts to source changes', async () => {
  const wrapper = await mountSuspended(UpdateMarkdown, { props: { source: '# Release\n\n[Docs](https://example.test)\n\n<script>alert(1)</script>' } })
  try {
    expect(wrapper.find('h2').text()).toBe('Release')
    expect(wrapper.find('script').exists()).toBe(false)
    expect(wrapper.find('a').attributes()).toMatchObject({ target: '_blank', rel: 'noopener noreferrer' })
    await wrapper.setProps({ source: '## 中文正文\n\n- 第一项' })
    expect(wrapper.find('h3').text()).toBe('中文正文')
    expect(wrapper.find('li').text()).toBe('第一项')
    expect(wrapper.find('a').exists()).toBe(false)
  } finally { wrapper.unmount() }
})
