import { ApplicationRef, NgModule, ComponentFactoryResolver } from '@angular/core'
import { BrowserModule } from '@angular/platform-browser'
import { ToastrModule } from 'ngx-toastr'

@NgModule({
    imports: [
        BrowserModule,
        ToastrModule.forRoot({
            positionClass: 'toast-bottom-center',
            toastClass: 'toast',
            preventDuplicates: true,
            extendedTimeOut: 1000,
        }),
    ],
})
export class AppModule {
    constructor (
        private appRef: ApplicationRef,
        private componentFactoryResolver: ComponentFactoryResolver,
    ) { }

    ngDoBootstrap () {
        const bootstrapComponents = (window as any)['bootstrapComponents'] || []
        if (bootstrapComponents.length === 0) {
            console.error('No bootstrap components found!')
            return
        }

        (window as any)['requestAnimationFrame'] = window[window['Zone'].__symbol__('requestAnimationFrame')]

        const componentDef = bootstrapComponents[0]
        const factory = this.componentFactoryResolver.resolveComponentFactory(componentDef)
        this.appRef.bootstrap(factory)
    }
}
