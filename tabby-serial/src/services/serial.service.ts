import { Injectable, Injector } from '@angular/core'
import { HostAppService, PartialProfile, Platform, ProfilesService } from 'tabby-core'
import WSABinding from 'serialport-binding-webserialapi'
import { SerialPortInfo, SerialProfile } from '../api'
import { SerialTabComponent } from '../components/serialTab.component'

@Injectable({ providedIn: 'root' })
export class SerialService {
    private constructor (
        private injector: Injector,
        private hostApp: HostAppService,


    ) { }

    detectBinding() { return this.hostApp.platform === Platform.Web ? WSABinding : null; }

    async listPorts (): Promise<SerialPortInfo[]> {
        try {
            if (this.hostApp.platform === Platform.Web) {
                return (await (this.detectBinding() as any).list()).map((x: any) => ({
                    name: x.path,
                    description: `${x.manufacturer ?? ''} ${x.serialNumber ?? ''}`.trim() || undefined,
                }))
            } else {
                const result = await window['require']('electron').ipcRenderer.invoke('serial:listPorts');
                return (result.ports || []).map((x: any) => ({
                    name: x.name,
                    description: `${x.manufacturer ?? ''} ${x.serialNumber ?? ''}`.trim() || undefined,
                }));
            }
        } catch (err) {
            console.error('Failed to list serial ports', err);
            return [];
        }
    }

    quickConnect (query: string): Promise<SerialTabComponent|null> {
        let path = query
        let baudrate = 115200
        if (query.includes('@')) {
            baudrate = parseInt(path.split('@')[1])
            path = path.split('@')[0]
        }
        const profile: PartialProfile<SerialProfile> = {
            name: query,
            type: 'serial',
            options: {
                port: path,
                baudrate: baudrate,
            },
        }
        window.localStorage.lastSerialConnection = JSON.stringify(profile)
        return this.injector.get(ProfilesService).openNewTabForProfile(profile) as Promise<SerialTabComponent|null>
    }
}
