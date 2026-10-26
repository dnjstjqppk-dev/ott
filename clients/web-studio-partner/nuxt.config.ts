// clients/web-studio-partner/nuxt.config.ts
export default defineNuxtConfig({
  compatibilityDate: '2024-09-23',
  ssr: false,
  devtools: { enabled: false },
  nitro: {
    preset: process.env.NITRO_PRESET || 'cloudflare-module',
    compatibilityDate: '2024-09-23'
  },
  modules: ['@pinia/nuxt', '@nuxtjs/tailwindcss'],
  optimization: {
    treeShake: {
      composables: {
        client: {},
        server: {}
      }
    }
  },
  typescript: {
    strict: true,
    typeCheck: false
  },
  runtimeConfig: {
    public: {
      apiGatewayUrl: process.env.NUXT_PUBLIC_API_GATEWAY_URL || 'http://localhost:8080'
    }
  }
})
