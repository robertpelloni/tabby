import { ApplicationRef, NgModule } from '@angular/core'
import { BrowserModule } from '@angular/platform-browser'
import { ToastrModule } from 'ngx-toastr'

export function createRootModule (plugins: any[]) {
    @NgModule({
        imports: [
            BrowserModule,
            ...plugins,
            ToastrModule.forRoot({
                positionClass: 'toast-bottom-center',
                toastClass: 'toast',
                preventDuplicates: true,
                extendedTimeOut: 1000,
            }),
        ],
    })
    class RootModule {
        constructor (private appRef: ApplicationRef) { }

        ngDoBootstrap () {
            const bootstrap = plugins.filter(x => x.bootstrap).map(x => x.bootstrap)
            if (bootstrap.length > 0) {
                (window as any)['requestAnimationFrame'] = window[window['Zone'].__symbol__('requestAnimationFrame')]
                this.appRef.bootstrap(bootstrap[0])
            }
        }
    }
    return RootModule
}
