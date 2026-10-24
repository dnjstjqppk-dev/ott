// clients/web-studio-partner/nuxt.config.ts
export default defineNuxtConfig({
  ssr: false,
  devtools: { enabled: false },
  modules: ['@pinia/nuxt', '@nuxtjs/tailwindcss'],
  typescript: {
    strict: true,
    typeCheck: true
  },
  runtimeConfig: {
    public: {
      apiGatewayUrl: process.env.NUXT_PUBLIC_API_GATEWAY_URL || 'http://localhost:8080'
    }
  }
})

