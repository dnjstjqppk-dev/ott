import { defineStore } from 'pinia';
import type { ProfileDto } from '~/types/netflix';

interface SessionState {
  accessToken: string;
  refreshToken: string;
  userId: string;
  deviceId: string;
  activeProfile: ProfileDto | null;
  profiles: ProfileDto[];
}

export const useSessionStore = defineStore('nf-session', {
  state: (): SessionState => {
    if (import.meta.client) {
      try {
        const raw = localStorage.getItem('nf-session');
        if (raw) return JSON.parse(raw) as SessionState;
      } catch {
        // ignore corrupted storage
      }
    }
    return {
      accessToken: '',
      refreshToken: '',
      userId: '',
      deviceId: '',
      activeProfile: null,
      profiles: []
    };
  },
  getters: {
    isAuthenticated: (state): boolean => state.accessToken.length > 0,
    authHeaders: (state): Record<string, string> => ({
      Authorization: `Bearer ${state.accessToken}`,
      'X-Netflix-Profile-Id': state.activeProfile?.id ?? ''
    })
  },
  actions: {
    async loginWithCredentials(email: string, password: string, fingerprint: string): Promise<void> {
      const config = useRuntimeConfig();
      const res = await $fetch<{
        access_token: string;
        refresh_token: string;
        user_id: string;
        device_id: string;
      }>(`${config.public.apiGatewayUrl}/api/v1/auth/login`, {
        method: 'POST',
        body: {
          email,
          password,
          device_fingerprint: fingerprint,
          device_name: navigator.userAgent.slice(0, 64),
          platform: 'WEB',
          drm_security_level: 'L1'
        }
      });
      this.accessToken = res.access_token;
      this.refreshToken = res.refresh_token;
      this.userId = res.user_id;
      this.deviceId = res.device_id;
      await this.fetchProfiles();
    },

    async fetchProfiles(): Promise<void> {
      const config = useRuntimeConfig();
      const res = await $fetch<{ profiles: ProfileDto[] }>(
        `${config.public.apiGatewayUrl}/api/v1/profiles`,
        { headers: this.authHeaders }
      );
      this.profiles = res.profiles;
if (import.meta.client) localStorage.setItem('nf-session', JSON.stringify(this.$state));
    },

    async selectProfile(profile: ProfileDto, rawPinCode?: string): Promise<boolean> {
      if (profile.has_pin) {
        const config = useRuntimeConfig();
        const verify = await $fetch<{ verified: boolean; profile_scoped_jwt: string }>(
          `${config.public.apiGatewayUrl}/api/v1/profiles/${profile.id}/verify-pin`,
          {
            method: 'POST',
            headers: this.authHeaders,
            body: { raw_pin_code: rawPinCode ?? '' }
          }
        );
        if (!verify.verified) {
          return false;
        }
        this.accessToken = verify.profile_scoped_jwt;
      }
      this.activeProfile = profile;
if (import.meta.client) localStorage.setItem('nf-session', JSON.stringify(this.$state));
      return true;
    },

    async logout(): Promise<void> {
      this.accessToken = '';
      this.refreshToken = '';
      this.userId = '';
      this.deviceId = '';
      this.activeProfile = null;
      this.profiles = [];
if (import.meta.client) localStorage.removeItem('nf-session');
    }
  }
});

