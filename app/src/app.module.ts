import { ApplicationRef, NgModule } from '@angular/core'
import { BrowserModule } from '@angular/platform-browser'
import { ToastrModule } from 'ngx-toastr'

<<<<<<< HEAD
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
    ]

    const bootstrap = [
        ...plugins.filter(x => x.bootstrap).map(x => x.bootstrap),
    ]

    if (bootstrap.length === 0) {
        throw new Error('Did not find any bootstrap components. Are there any plugins installed?')
    }

=======
export function createRootModule (plugins: any[]) {
>>>>>>> jules-1407546259735951285-590dfa06
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
