import { render } from '@testing-library/react'
/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/
import type { ReactNode } from 'react'
import { describe, expect, test, vi } from 'vitest'

vi.mock('@tanstack/react-router', () => ({
  Link: ({ children }: { children?: ReactNode }) => <a>{children}</a>,
}))

vi.mock('@/components/animate-in-view', () => ({
  AnimateInView: ({ children }: { children?: ReactNode }) => (
    <div>{children}</div>
  ),
}))

vi.mock('@/components/ui/accordion', () => ({
  Accordion: ({ children }: { children?: ReactNode }) => <div>{children}</div>,
  AccordionContent: ({ children }: { children?: ReactNode }) => (
    <div>{children}</div>
  ),
  AccordionItem: ({ children }: { children?: ReactNode }) => (
    <div>{children}</div>
  ),
  AccordionTrigger: ({ children }: { children?: ReactNode }) => (
    <button type='button'>{children}</button>
  ),
}))

vi.mock('@/components/ui/button', () => ({
  Button: ({ children }: { children?: ReactNode }) => (
    <button type='button'>{children}</button>
  ),
}))

vi.mock('@/hooks/use-status', () => ({
  useStatus: () => ({ status: null }),
}))

vi.mock('../../hero-terminal-demo', () => ({
  HeroTerminalDemo: () => <div />,
}))

const { Hero } = await import('../hero')
const { ApiMixHomeContent } = await import('../api-mix-home-content')

describe('API MIX home layout', () => {
  test('keeps the capability strip close to the hero content', () => {
    const { container } = render(<Hero isAuthenticated={false} />)
    const hero = container.querySelector('section')

    expect(hero).toHaveClass('pb-16', 'md:pb-20', 'lg:pb-16')
    expect(hero).not.toHaveClass('min-h-[100dvh]')

    const { container: contentContainer } = render(
      <ApiMixHomeContent isAuthenticated={false} />
    )
    const capabilityStrip = contentContainer.querySelector('main > section')

    expect(capabilityStrip).toHaveClass('-mt-12', 'pb-12', 'md:pb-16')
  })

  test('uses a compact final call to action', () => {
    const { container } = render(<ApiMixHomeContent isAuthenticated={false} />)
    const sections = container.querySelectorAll('main > section')
    const finalCta = sections.item(sections.length - 1)

    expect(finalCta).toHaveClass('py-14', 'md:py-16')
    expect(finalCta).not.toHaveClass('py-24', 'md:py-32')
  })
})
