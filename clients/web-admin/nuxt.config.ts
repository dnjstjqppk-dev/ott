// clients/web-admin/nuxt.config.ts
declare const process: any;

export default defineNuxtConfig({
  compatibilityDate: '2024-09-23',
  ssr: true,
  devtools: { enabled: false },
  nitro: {
    preset: process?.env?.NITRO_PRESET || 'cloudflare-module',
    compatibilityDate: '2024-09-23'
  } as any,
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
      apiGatewayUrl: process?.env?.NUXT_PUBLIC_API_GATEWAY_URL || ''
    }
  },
  app: {
    head: {
      title: 'Netflix Admin & SRE Control Plane',
      meta: [
        { charset: 'utf-8' },
        { name: 'viewport', content: 'width=device-width, initial-scale=1' }
      ]
    }
  }
});
