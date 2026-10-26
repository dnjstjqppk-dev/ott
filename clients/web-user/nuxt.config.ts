// clients/web-user/nuxt.config.ts
export default defineNuxtConfig({
  compatibilityDate: '2024-09-23',
  ssr: true,
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
      apiGatewayUrl: process.env.NUXT_PUBLIC_API_GATEWAY_URL || 'http://localhost:8080',
      drmLicenseUrl: process.env.NUXT_PUBLIC_DRM_LICENSE_URL || 'http://localhost:8080/api/v1/playback/drm-license'
    }
  },
  app: {
    head: {
      title: 'Netflix',
      meta: [
        { charset: 'utf-8' },
        { name: 'viewport', content: 'width=device-width, initial-scale=1' }
      ]
    }
  }
})
