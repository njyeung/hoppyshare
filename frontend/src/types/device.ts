export interface DeviceSettings {
  nickname: string;
  enabled: boolean;
  auto_copy: boolean;
  light_animations: boolean;
  cache_time: number;
  muted: boolean;
  send_to_self: boolean;
  auto_ble: boolean;
  startup: boolean;
  destroy: boolean;
}

export interface Device {
  deviceid: string;
  settings: DeviceSettings;
}