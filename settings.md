This document defines the standardized device settings schema used across all components of HoppyShare:
- **Python Backend** (Lambda) - Generates default settings
- **Frontend** (React/TypeScript) - Displays and updates settings  
- **Desktop Client** (Go) - Parses and applies settings

## Settings Schema

### JSON Structure
```json
{
  "deviceid": "75c5896d-ff07-4131-b486-bc8af70e6b03",
  "settings": {
    "nickname": "My Device",
    "enabled": true,
    "auto_copy": false,
    "auto_paste": false,
    "cache_time": 30,
    "muted": false,
    "send_to_self": true,
    "auto_ble": true,
    "startup": true,
    "destroy": false
  }
}
```

### Setting Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `nickname` | `string` | `"Unnamed Device"` | Human-readable device name displayed in UI |
| `enabled` | `boolean` | `true` | Master switch - when false, device ignores all messages |
| `auto_copy` | `boolean` | `false` | Automatically copy received messages to clipboard |
| `auto_paste` | `boolean` | `false` | Automatically paste after auto_copy (requires auto_copy: true) |
| `cache_time` | `number` | `30` | Time in seconds messages remain accessible (max 300s) |
| `muted` | `boolean` | `false` | Disable all notification sounds |
| `send_to_self` | `boolean` | `true` | Allow receiving messages from the same device |
| `auto_ble` | `boolean` | `true` | Automatically enable BLE when network connection is lost |
| `startup` | `boolean` | `true` | Launch application automatically on system boot |
| `destroy` | `boolean` | `false` | Self-destruct flag - quit and remove application |

## Implementation Notes

### Default Values
All three components should use these exact default values:

**Python (Lambda)**:
```python
"settings": {
    "nickname": "Unnamed Device",
    "enabled": True,
    "auto_copy": False,
    "auto_paste": False,
    "cache_time": 30,
    "muted": False,
    "send_to_self": True,
    "auto_ble": True,
    "startup": True,
    "destroy": False
}
```

**TypeScript (Frontend)**:
```typescript
interface DeviceSettings {
  nickname: string;      // "Unnamed Device"
  enabled: boolean;      // true
  auto_copy: boolean;    // false
  auto_paste: boolean;   // false
  cache_time: number;    // 30
  muted: boolean;        // false
  send_to_self: boolean; // true
  auto_ble: boolean;     // true
  startup: boolean;      // true
  destroy: boolean;      // false
}
```

**Go (Desktop Client)**:
```go
type Settings struct {
    Nickname   string // "Unnamed Device"
    Enabled    bool   // true
    AutoCopy   bool   // false (maps to auto_copy)
    AutoPaste  bool   // false (maps to auto_paste) 
    CacheTime  int    // 30 (maps to cache_time)
    Muted      bool   // false
    SendToSelf bool   // true (maps to send_to_self)
    AutoBLE    bool   // true (maps to auto_ble)
    Startup    bool   // true
    Destroy    bool   // false
}
```

### Field Name Mapping
The Go client uses PascalCase while JSON uses snake_case:

| JSON Field | Go Field |
|------------|----------|
| `auto_copy` | `AutoCopy` |
| `auto_paste` | `AutoPaste` |
| `cache_time` | `CacheTime` |
| `send_to_self` | `SendToSelf` |
| `auto_ble` | `AutoBLE` |

### Validation Rules
- `cache_time`: Must be between 1 and 300 seconds
- `nickname`: fallback to "Unnamed Device"
- `auto_paste`: Only functions when `auto_copy` is true

### Special Behaviors
- **destroy**: When set to true, causes the desktop client to quit and self-remove
- **startup**: Changes to this setting immediately update system startup registry/autostart
- **auto_ble**: Controls whether BLE automatically activates on network loss
- **enabled**: Master switch that disables all message processing when false
