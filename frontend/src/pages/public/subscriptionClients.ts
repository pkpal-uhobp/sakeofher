export type DeviceKey = 'android' | 'ios' | 'windows' | 'macos' | 'linux'
export type ClientCore = 'xray' | 'mihomo' | 'singbox'

export interface SubscriptionDevice {
  key: DeviceKey
  label: string
}

export interface SubscriptionClientCard {
  device: DeviceKey
  id: string
  name: string
  core: ClientCore
  description: string
  featured: boolean
  hwid: boolean
  download: string
  open: string
}

export const SUBSCRIPTION_DEVICES: SubscriptionDevice[] = [
  { key: 'android', label: 'Android' },
  { key: 'ios', label: 'iOS' },
  { key: 'windows', label: 'Windows' },
  { key: 'macos', label: 'macOS' },
  { key: 'linux', label: 'Linux' },
]

// Pool like on the Remnawave subscription page shown by the user,
// plus requested iOS clients: V2RayTun and Shadowrocket.
const clients = [
  {
    id: 'flclashx',
    name: 'FlClashX',
    core: 'mihomo',
    devices: ['android', 'windows', 'macos', 'linux'],
    featured: true,
    hwid: true,
    description: 'Mihomo/Clash-клиент из пула Remnawave.',
    downloads: {
      android: 'https://github.com/pluralplay/FlClashX/releases/latest',
      windows: 'https://github.com/pluralplay/FlClashX/releases/latest',
      macos: 'https://github.com/pluralplay/FlClashX/releases/latest',
      linux: 'https://github.com/pluralplay/FlClashX/releases/latest',
    },
  },
  {
    id: 'happ',
    name: 'Happ',
    core: 'xray',
    devices: ['android', 'ios', 'macos', 'windows', 'linux'],
    featured: true,
    hwid: true,
    description: 'Xray-клиент для Android, iOS, macOS, Windows и Linux.',
    downloads: {
      android: 'https://play.google.com/store/apps/details?id=com.happproxy',
      ios: 'https://apps.apple.com/us/app/happ-proxy-utility/id6504287215',
      macos: 'https://apps.apple.com/us/app/happ-proxy-utility/id6504287215',
      windows: 'https://github.com/Happ-proxy/happ-desktop/releases/latest',
      linux: 'https://github.com/Happ-proxy/happ-desktop/releases/latest',
    },
  },
  {
    id: 'koalaclash',
    name: 'Koala Clash',
    core: 'mihomo',
    devices: ['windows', 'macos', 'linux'],
    featured: true,
    hwid: true,
    description: 'Desktop Mihomo-клиент из пула Remnawave.',
    downloads: {
      windows: 'https://github.com/coolcoala/clash-verge-rev-lite/releases/latest',
      macos: 'https://github.com/coolcoala/clash-verge-rev-lite/releases/latest',
      linux: 'https://github.com/coolcoala/clash-verge-rev-lite/releases/latest',
    },
  },
  {
    id: 'clash-verge',
    name: 'Clash Verge',
    core: 'mihomo',
    devices: ['windows', 'macos', 'linux'],
    featured: false,
    hwid: false,
    description: 'Desktop Mihomo/Clash-клиент.',
    downloads: {
      windows: 'https://github.com/clash-verge-rev/clash-verge-rev/releases/latest',
      macos: 'https://github.com/clash-verge-rev/clash-verge-rev/releases/latest',
      linux: 'https://github.com/clash-verge-rev/clash-verge-rev/releases/latest',
    },
  },
  {
    id: 'flclash',
    name: 'FlClash',
    core: 'mihomo',
    devices: ['android', 'windows', 'macos', 'linux'],
    featured: false,
    hwid: false,
    description: 'Кроссплатформенный Mihomo-клиент.',
    downloads: {
      android: 'https://github.com/chen08209/FlClash/releases/latest',
      windows: 'https://github.com/chen08209/FlClash/releases/latest',
      macos: 'https://github.com/chen08209/FlClash/releases/latest',
      linux: 'https://github.com/chen08209/FlClash/releases/latest',
    },
  },
  {
    id: 'hiddify',
    name: 'Hiddify',
    core: 'singbox',
    devices: ['android', 'ios', 'macos', 'windows', 'linux'],
    featured: false,
    hwid: false,
    description: 'Кроссплатформенный sing-box-клиент.',
    downloads: {
      android: 'https://play.google.com/store/apps/details?id=app.hiddify.com',
      ios: 'https://apps.apple.com/us/app/hiddify-proxy-vpn/id6596777532',
      macos: 'https://github.com/hiddify/hiddify-app/releases',
      windows: 'https://github.com/hiddify/hiddify-app/releases',
      linux: 'https://github.com/hiddify/hiddify-app/releases',
    },
  },
  {
    id: 'v2raytun',
    name: 'V2RayTun',
    core: 'xray',
    devices: ['ios'],
    featured: false,
    hwid: false,
    description: 'iOS-клиент с поддержкой Xray и deep link импорта.',
    downloads: {
      ios: 'https://apps.apple.com/us/app/v2raytun/id6476628951',
    },
  },
  {
    id: 'shadowrocket',
    name: 'Shadowrocket',
    core: 'xray',
    devices: ['ios'],
    featured: false,
    hwid: false,
    description: 'Популярный iOS proxy-клиент.',
    downloads: {
      ios: 'https://apps.apple.com/us/app/shadowrocket/id932747118',
    },
  },
] satisfies Array<{
  id: string
  name: string
  core: ClientCore
  devices: DeviceKey[]
  description: string
  featured: boolean
  hwid: boolean
  downloads: Partial<Record<DeviceKey, string>>
}>

export function buildSubscriptionApps(params: {
  subscriptionURL: string
}): SubscriptionClientCard[] {
  const raw = params.subscriptionURL
  const encoded = encodeURIComponent(raw)
  const encodedURI = encodeURI(raw)
  const base64 = toBase64(raw)
  const name = encodeURIComponent('SakeOfHer')

  return clients
    .flatMap((client) =>
      client.devices.map((device) => ({
        device,
        id: client.id,
        name: client.name,
        core: client.core,
        description: client.description,
        featured: client.featured,
        hwid: client.hwid,
        download: client.downloads[device] || raw,
        open: buildOpenURL(client.id, client.core, {
          raw,
          encoded,
          encodedURI,
          base64,
          name,
        }),
      })),
    )
    .sort((a, b) => Number(b.featured) - Number(a.featured) || a.name.localeCompare(b.name))
}

function buildOpenURL(
  id: string,
  core: ClientCore,
  link: {
    raw: string
    encoded: string
    encodedURI: string
    base64: string
    name: string
  },
): string {
  switch (id) {
    case 'happ':
      return `hiddify://install-sub/?url=${link.encoded}`

    case 'hiddify':
      return `hiddify://import/${link.encodedURI}#${link.name}`

    case 'v2raytun':
      return `v2raytun://import/${link.raw}`

    case 'shadowrocket':
      return `shadowrocket://add/sub://${link.base64}?remarks=${link.name}`

    case 'flclashx':
    case 'flclash':
    case 'clash-verge':
    case 'koalaclash':
      return `clash://install-config?url=${link.encoded}&name=${link.name}`

    default:
      if (core === 'mihomo') {
        return `clash://install-config?url=${link.encoded}&name=${link.name}`
      }

      return link.raw
  }
}

function toBase64(value: string): string {
  return btoa(unescape(encodeURIComponent(value)))
}
