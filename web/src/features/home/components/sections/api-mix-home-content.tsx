/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.
*/
import { Link } from '@tanstack/react-router'
import {
  ArrowRight,
  Check,
  Code,
  Cpu,
  Image,
  Key,
  LayoutDashboard,
  MessageSquare,
  Mic,
  Play,
  ShieldCheck,
  Sparkles,
  Workflow,
  Zap,
} from 'lucide-react'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'

import { AnimateInView } from '@/components/animate-in-view'
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from '@/components/ui/accordion'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'

interface ApiMixHomeContentProps {
  isAuthenticated: boolean
}

const providers = [
  'OpenAI',
  'Anthropic',
  'Google',
  'DeepSeek',
  'Qwen',
  'Mistral',
  'Llama',
  'Cohere',
]

const modelCategories = [
  {
    id: 'chat',
    label: 'Conversation',
    icon: MessageSquare,
    models: ['GPT-4o', 'Claude', 'Gemini'],
  },
  {
    id: 'vision',
    label: 'Vision',
    icon: Image,
    models: ['GPT-4o', 'Gemini', 'Qwen-VL'],
  },
  {
    id: 'audio',
    label: 'Audio',
    icon: Mic,
    models: ['Whisper', 'GPT-4o Audio', 'Gemini'],
  },
  {
    id: 'video',
    label: 'Video',
    icon: Play,
    models: ['Sora', 'Veo', 'Kling'],
  },
]

export function ApiMixHomeContent({ isAuthenticated }: ApiMixHomeContentProps) {
  const { t } = useTranslation()
  const [activeCategory, setActiveCategory] = useState('chat')
  const activeModels =
    modelCategories.find((category) => category.id === activeCategory) ??
    modelCategories[0]

  const actionRoute = isAuthenticated ? '/dashboard' : '/sign-up'
  const actionLabel = isAuthenticated ? 'Go to Dashboard' : 'Get Started'

  const features = [
    {
      image: '/feature-unified-api.png',
      icon: Sparkles,
      title: 'One interface, many providers',
      description:
        'Use a consistent API surface while routing requests to the providers configured for your workspace.',
    },
    {
      image: '/feature-model-routing.png',
      icon: Workflow,
      title: 'Routing under your control',
      description:
        'Organize channels and models to match the access, reliability, and cost rules your team needs.',
    },
    {
      image: '/feature-cost-control.png',
      icon: ShieldCheck,
      title: 'Usage you can inspect',
      description:
        'Review token usage, spending, and request activity from the same operational workspace.',
    },
    {
      image: '/feature-async-tasks.png',
      icon: Zap,
      title: 'Built for more than chat',
      description:
        'Bring chat, image, audio, and asynchronous model workflows together behind one gateway.',
    },
  ]

  const steps = [
    {
      number: '01',
      icon: Key,
      title: 'Create an API key',
      description:
        'Generate a key from your account and define the access scope for each application.',
    },
    {
      number: '02',
      icon: Workflow,
      title: 'Select your routes',
      description:
        'Connect the channels and model groups that match your product requirements.',
    },
    {
      number: '03',
      icon: Code,
      title: 'Ship with one endpoint',
      description:
        'Use compatible API routes to bring model calls into your existing application workflow.',
    },
  ]

  const dashboardCapabilities = [
    'Monitor request and usage activity',
    'Manage API keys and access policies',
    'Review channels, models, and billing data',
  ]

  const faqs = [
    {
      question: 'What can I use API MIX for?',
      answer:
        'API MIX provides a unified gateway for the AI models and providers configured by your workspace. Available capabilities depend on your administrator settings and account permissions.',
    },
    {
      question: 'Can I use my existing API integration?',
      answer:
        'The platform supports compatible API routes for common AI application workflows. Confirm the available models and route settings in your dashboard before switching production traffic.',
    },
    {
      question: 'Where can I check pricing and usage?',
      answer:
        'Your current pricing and usage details are available from the pricing page and your account dashboard.',
    },
  ]

  return (
    <main className='relative overflow-hidden bg-[#f7faff] text-[#14233d] dark:bg-[#090d16] dark:text-[#f5f7fa]'>
      <div
        aria-hidden
        className='pointer-events-none absolute inset-0 opacity-55 dark:opacity-100'
        style={{
          backgroundImage:
            'linear-gradient(rgba(67, 99, 148, .07) 1px, transparent 1px), linear-gradient(90deg, rgba(67, 99, 148, .07) 1px, transparent 1px)',
          backgroundSize: '72px 72px',
        }}
      />

      <section className='relative z-10 px-5 pb-12 sm:px-8 md:pb-16'>
        <AnimateInView animation='scale-in' className='mx-auto max-w-6xl'>
          <div className='grid border-y border-[#cfdaea] md:grid-cols-4 dark:border-[#263247]'>
            {[
              [
                Cpu,
                'Unified gateway',
                'A single endpoint for configured AI services',
              ],
              [
                Workflow,
                'Flexible routing',
                'Keep routing decisions in your workspace',
              ],
              [
                LayoutDashboard,
                'Operational visibility',
                'Inspect usage and request activity',
              ],
              [
                ShieldCheck,
                'Account controls',
                'Manage access with API keys and roles',
              ],
            ].map(([Icon, label, description], index) => {
              const StatIcon = Icon as typeof Cpu
              return (
                <div
                  key={label as string}
                  className={cn(
                    'flex items-start gap-3 px-1 py-5 md:px-6 md:py-6',
                    index > 0 &&
                      'border-t border-[#dbe4f1] md:border-l md:border-t-0 dark:border-[#202b3d]'
                  )}
                >
                  <div className='mt-0.5 flex size-8 shrink-0 items-center justify-center bg-[#eaf3ff] text-[#2f80ed] dark:bg-[#142b49] dark:text-[#75b5ff]'>
                    <StatIcon className='size-4' />
                  </div>
                  <div className='min-w-0'>
                    <p className='text-sm font-semibold'>
                      {t(label as string)}
                    </p>
                    <p className='mt-1 text-xs leading-5 text-[#64738c] dark:text-[#8594ae]'>
                      {t(description as string)}
                    </p>
                  </div>
                </div>
              )
            })}
          </div>
        </AnimateInView>
      </section>

      <section className='relative z-10 px-5 py-16 sm:px-8 md:py-24'>
        <div className='mx-auto max-w-6xl text-center'>
          <AnimateInView>
            <p className='text-xs font-semibold tracking-[0.18em] text-[#2f80ed] uppercase dark:text-[#67a7ff]'>
              {t('Provider ecosystem')}
            </p>
            <h2 className='mt-4 text-3xl font-semibold tracking-[0] md:text-4xl'>
              {t('Connect the models your product needs')}
            </h2>
            <p className='mx-auto mt-4 max-w-2xl text-sm leading-7 text-[#64738c] md:text-base dark:text-[#8d9bb3]'>
              {t(
                'Keep provider choice flexible while presenting your applications with one dependable integration surface.'
              )}
            </p>
          </AnimateInView>
          <div className='mt-12 flex flex-wrap items-center justify-center gap-x-10 gap-y-6 md:gap-x-14'>
            {providers.map((provider, index) => (
              <AnimateInView
                key={provider}
                delay={index * 55}
                animation='fade-up'
              >
                <span className='text-base font-semibold text-[#70809a] transition-colors hover:text-[#2f80ed] md:text-lg dark:text-[#53627c] dark:hover:text-[#a6caff]'>
                  {provider}
                </span>
              </AnimateInView>
            ))}
          </div>
        </div>
      </section>

      <section className='relative z-10 px-5 py-20 sm:px-8 md:py-28'>
        <div className='mx-auto max-w-6xl'>
          <AnimateInView className='mx-auto max-w-2xl text-center'>
            <p className='text-xs font-semibold tracking-[0.18em] text-[#2f80ed] uppercase dark:text-[#67a7ff]'>
              {t('Platform capabilities')}
            </p>
            <h2 className='mt-4 text-3xl font-semibold tracking-[0] md:text-4xl'>
              {t('A focused layer for AI delivery')}
            </h2>
            <p className='mt-4 text-sm leading-7 text-[#64738c] md:text-base dark:text-[#8d9bb3]'>
              {t(
                'Bring provider management, application access, and operational insight into one environment.'
              )}
            </p>
          </AnimateInView>
          <div className='mt-14 grid gap-5 md:grid-cols-2'>
            {features.map((feature, index) => {
              const Icon = feature.icon
              return (
                <AnimateInView
                  key={feature.title}
                  delay={index * 80}
                  animation='fade-up'
                >
                  <article className='group h-full overflow-hidden border border-[#d3deec] bg-white transition-[transform,border-color,box-shadow] duration-300 hover:-translate-y-1 hover:border-[#79b1ef] hover:shadow-[0_18px_50px_rgba(39,104,183,.14)] dark:border-[#263247] dark:bg-[#0d1320] dark:hover:border-[#3a6ca9]'>
                    <div className='h-48 overflow-hidden border-b border-[#e0e8f3] bg-[#edf4fd] dark:border-[#202b3d] dark:bg-[#0a101b]'>
                      <img
                        src={feature.image}
                        alt={t(feature.title)}
                        className='size-full object-cover opacity-90 transition duration-500 group-hover:scale-[1.03] group-hover:opacity-100'
                        loading='lazy'
                      />
                    </div>
                    <div className='p-6 md:p-7'>
                      <div className='flex size-11 items-center justify-center bg-[#eaf3ff] text-[#2f80ed] dark:bg-[#1b3a64]/50 dark:text-[#76b4ff]'>
                        <Icon className='size-5' />
                      </div>
                      <h3 className='mt-5 text-lg font-semibold'>
                        {t(feature.title)}
                      </h3>
                      <p className='mt-3 text-sm leading-7 text-[#64738c] dark:text-[#8d9bb3]'>
                        {t(feature.description)}
                      </p>
                    </div>
                  </article>
                </AnimateInView>
              )
            })}
          </div>
        </div>
      </section>

      <section className='relative z-10 border-y border-[#dbe5f2] bg-[#eff5fc] px-5 py-20 sm:px-8 md:py-28 dark:border-[#202b3d] dark:bg-[#0d1320]'>
        <div className='mx-auto max-w-6xl'>
          <AnimateInView className='mx-auto max-w-2xl text-center'>
            <p className='text-xs font-semibold tracking-[0.18em] text-[#2f80ed] uppercase dark:text-[#67a7ff]'>
              {t('Getting started')}
            </p>
            <h2 className='mt-4 text-3xl font-semibold tracking-[0] md:text-4xl'>
              {t('Three steps to your first request')}
            </h2>
            <p className='mt-4 text-sm leading-7 text-[#64738c] md:text-base dark:text-[#8d9bb3]'>
              {t(
                'Set up only what your application needs, then evolve your routing as the product grows.'
              )}
            </p>
          </AnimateInView>
          <div className='relative mt-14 grid gap-10 md:grid-cols-3 md:gap-8'>
            <div className='absolute top-12 right-[18%] left-[18%] hidden border-t border-dashed border-[#7fb4ed] md:block dark:border-[#315985]' />
            {steps.map((step, index) => {
              const Icon = step.icon
              return (
                <AnimateInView
                  key={step.number}
                  delay={index * 110}
                  animation='fade-up'
                >
                  <div className='relative text-center'>
                    <span className='font-mono text-5xl font-bold text-[#2f80ed]/25 dark:text-[#6399dc]/25'>
                      {step.number}
                    </span>
                    <div className='mx-auto mt-5 flex size-16 items-center justify-center border border-[#b8d1ed] bg-white text-[#2f80ed] shadow-sm dark:border-[#315273] dark:bg-[#101b2b] dark:text-[#76b4ff]'>
                      <Icon className='size-7' />
                    </div>
                    <h3 className='mt-6 text-lg font-semibold'>
                      {t(step.title)}
                    </h3>
                    <p className='mx-auto mt-3 max-w-xs text-sm leading-7 text-[#64738c] dark:text-[#8d9bb3]'>
                      {t(step.description)}
                    </p>
                  </div>
                </AnimateInView>
              )
            })}
          </div>
        </div>
      </section>

      <section className='relative z-10 px-5 py-20 sm:px-8 md:py-28'>
        <div className='mx-auto grid max-w-6xl items-center gap-12 lg:grid-cols-5'>
          <div className='lg:col-span-2'>
            <AnimateInView>
              <p className='text-xs font-semibold tracking-[0.18em] text-[#2f80ed] uppercase dark:text-[#67a7ff]'>
                {t('Developer workspace')}
              </p>
              <h2 className='mt-4 text-3xl font-semibold tracking-[0] md:text-4xl'>
                {t('See the system behind each request')}
              </h2>
              <p className='mt-4 text-sm leading-7 text-[#64738c] md:text-base dark:text-[#8d9bb3]'>
                {t(
                  'Use the dashboard to manage access, inspect operations, and keep your AI integration organized.'
                )}
              </p>
            </AnimateInView>
            <div className='mt-8 space-y-4'>
              {dashboardCapabilities.map((capability, index) => (
                <AnimateInView
                  key={capability}
                  delay={100 + index * 75}
                  animation='fade-up'
                >
                  <div className='flex items-center gap-3 text-sm text-[#31445f] dark:text-[#c5d1e3]'>
                    <Check className='size-5 shrink-0 text-[#1aa884] dark:text-[#44d7b6]' />
                    {t(capability)}
                  </div>
                </AnimateInView>
              ))}
            </div>
            <AnimateInView delay={330} animation='fade-up'>
              <Button
                variant='link'
                className='mt-8 h-auto px-0 text-[#2f80ed] dark:text-[#76b4ff]'
                render={<Link to='/dashboard' />}
              >
                {t('Open dashboard')}
                <ArrowRight className='ml-1.5 size-4' />
              </Button>
            </AnimateInView>
          </div>
          <AnimateInView animation='fade-left' className='lg:col-span-3'>
            <Link
              to='/dashboard'
              className='group block overflow-hidden border border-[#ccd9e9] bg-[#0d1320] shadow-[0_28px_70px_rgba(21,50,90,.2)] dark:border-[#263247]'
            >
              <div className='flex items-center gap-2 border-b border-[#25344c] px-4 py-3'>
                <span className='size-2 rounded-full bg-[#ef6b73]' />
                <span className='size-2 rounded-full bg-[#e2b85d]' />
                <span className='size-2 rounded-full bg-[#45c29a]' />
                <span className='ml-2 text-[10px] font-medium tracking-[0.12em] text-[#8292ab] uppercase'>
                  API MIX / {t('Dashboard')}
                </span>
              </div>
              <img
                src='/hero-dashboard-preview.png'
                alt={t('API MIX dashboard preview')}
                className='block aspect-[16/10] w-full object-cover transition duration-500 group-hover:scale-[1.015]'
                loading='lazy'
              />
            </Link>
          </AnimateInView>
        </div>
      </section>

      <section className='relative z-10 border-y border-[#dbe5f2] bg-[#eff5fc] px-5 py-20 sm:px-8 md:py-28 dark:border-[#202b3d] dark:bg-[#0d1320]'>
        <div className='mx-auto max-w-6xl'>
          <AnimateInView className='mx-auto max-w-2xl text-center'>
            <p className='text-xs font-semibold tracking-[0.18em] text-[#2f80ed] uppercase dark:text-[#67a7ff]'>
              {t('Model capabilities')}
            </p>
            <h2 className='mt-4 text-3xl font-semibold tracking-[0] md:text-4xl'>
              {t('Design for the modality you need')}
            </h2>
            <p className='mt-4 text-sm leading-7 text-[#64738c] md:text-base dark:text-[#8d9bb3]'>
              {t(
                'Explore the model categories exposed through the routes your workspace has configured.'
              )}
            </p>
          </AnimateInView>
          <div className='mx-auto mt-12 max-w-4xl'>
            <div className='grid grid-cols-2 border border-[#cbd9e9] bg-white p-1 sm:grid-cols-4 dark:border-[#2a3a51] dark:bg-[#0a101b]'>
              {modelCategories.map((category) => {
                const Icon = category.icon
                const isActive = category.id === activeCategory
                return (
                  <button
                    type='button'
                    key={category.id}
                    onClick={() => setActiveCategory(category.id)}
                    className={cn(
                      'flex min-h-12 items-center justify-center gap-2 px-3 text-xs font-semibold transition-colors',
                      isActive
                        ? 'bg-[#2f80ed] text-white'
                        : 'text-[#667792] hover:bg-[#eef5ff] hover:text-[#2f80ed] dark:text-[#8494ae] dark:hover:bg-[#14253c] dark:hover:text-[#9bc8ff]'
                    )}
                  >
                    <Icon className='size-4' />
                    {t(category.label)}
                  </button>
                )
              })}
            </div>
            <div className='mt-5 grid gap-5 md:grid-cols-3'>
              {activeModels.models.map((model, index) => (
                <div
                  key={model}
                  className='border border-[#cbd9e9] bg-white p-6 dark:border-[#2a3a51] dark:bg-[#101b2b]'
                >
                  <span className='font-mono text-xs text-[#2f80ed] dark:text-[#76b4ff]'>
                    0{index + 1}
                  </span>
                  <h3 className='mt-6 text-lg font-semibold'>{model}</h3>
                  <p className='mt-2 text-sm leading-6 text-[#64738c] dark:text-[#8d9bb3]'>
                    {t(
                      'Availability is managed by your configured model groups and account access.'
                    )}
                  </p>
                </div>
              ))}
            </div>
          </div>
        </div>
      </section>

      <section className='relative z-10 overflow-hidden bg-[#166bd0] px-5 py-20 text-white sm:px-8 md:py-28'>
        <div
          aria-hidden
          className='absolute inset-0 opacity-30'
          style={{
            backgroundImage:
              'radial-gradient(circle at 1px 1px, rgba(255,255,255,.38) 1px, transparent 0)',
            backgroundSize: '48px 48px',
          }}
        />
        <AnimateInView
          className='relative mx-auto max-w-3xl text-center'
          animation='scale-in'
        >
          <p className='text-xs font-semibold tracking-[0.18em] text-[#bfe1ff] uppercase'>
            {t('Transparent account controls')}
          </p>
          <h2 className='mt-4 text-3xl font-semibold tracking-[0] md:text-4xl'>
            {t('Pricing that stays in view')}
          </h2>
          <p className='mx-auto mt-4 max-w-xl text-sm leading-7 text-[#e4f1ff] md:text-base'>
            {t(
              'Review the plans and usage rules available to your account before you scale.'
            )}
          </p>
          <Button
            className='mt-9 h-12 bg-white px-6 font-semibold text-[#166bd0] hover:bg-[#edf6ff]'
            render={<Link to='/pricing' />}
          >
            {t('View Pricing')}
            <ArrowRight className='ml-1.5 size-4' />
          </Button>
        </AnimateInView>
      </section>

      <section className='relative z-10 px-5 py-20 sm:px-8 md:py-28'>
        <div className='mx-auto max-w-3xl'>
          <AnimateInView className='text-center'>
            <p className='text-xs font-semibold tracking-[0.18em] text-[#2f80ed] uppercase dark:text-[#67a7ff]'>
              {t('FAQ')}
            </p>
            <h2 className='mt-4 text-3xl font-semibold tracking-[0] md:text-4xl'>
              {t('Questions before you connect?')}
            </h2>
          </AnimateInView>
          <AnimateInView
            className='mt-10 border-y border-[#dbe5f2] dark:border-[#263247]'
            animation='fade-up'
          >
            <Accordion>
              {faqs.map((faq) => (
                <AccordionItem
                  key={faq.question}
                  value={faq.question}
                  className='border-[#dbe5f2] dark:border-[#263247]'
                >
                  <AccordionTrigger className='py-5 text-base hover:no-underline'>
                    {t(faq.question)}
                  </AccordionTrigger>
                  <AccordionContent className='max-w-2xl pb-5 text-sm leading-7 text-[#64738c] dark:text-[#8d9bb3]'>
                    {t(faq.answer)}
                  </AccordionContent>
                </AccordionItem>
              ))}
            </Accordion>
          </AnimateInView>
        </div>
      </section>

      <section className='relative z-10 overflow-hidden border-t border-[#dbe5f2] px-5 py-14 sm:px-8 md:py-16 dark:border-[#202b3d]'>
        <div
          aria-hidden
          className='absolute inset-0 opacity-50 dark:opacity-100'
          style={{
            backgroundImage:
              'radial-gradient(circle at 1px 1px, rgba(47,128,237,.22) 1px, transparent 0)',
            backgroundSize: '58px 58px',
          }}
        />
        <AnimateInView
          className='relative mx-auto max-w-3xl text-center'
          animation='scale-in'
        >
          <h2 className='text-4xl font-semibold tracking-[0] md:text-5xl'>
            {t('Build your AI layer with API MIX')}
          </h2>
          <p className='mx-auto mt-5 max-w-xl text-sm leading-7 text-[#64738c] md:text-base dark:text-[#8d9bb3]'>
            {t(
              'Create an account to configure your workspace and start connecting the models your product needs.'
            )}
          </p>
          <Button
            className='mt-9 h-12 bg-[#2f80ed] px-6 font-semibold text-white hover:bg-[#4392ff]'
            render={<Link to={actionRoute} />}
          >
            {t(actionLabel)}
            <ArrowRight className='ml-1.5 size-4' />
          </Button>
        </AnimateInView>
      </section>
    </main>
  )
}
