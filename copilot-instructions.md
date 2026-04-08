# Copilot Instructions

> This file provides instructions for GitHub Copilot when working with the Tabby project.
> **Read [docs/UNIVERSAL_LLM_INSTRUCTIONS.md](docs/UNIVERSAL_LLM_INSTRUCTIONS.md) for comprehensive project rules.**

## Quick Reference

### Project Type
Cross-platform terminal emulator (Electron + Angular + TypeScript)

### Key Conventions
- Angular 15 with Pug templates and SCSS
- Plugin architecture: each `tabby-*` directory is a plugin module
- Services use `@Injectable()`, Components use `@Component()`
- Config is YAML-based with platform-specific defaults
- Version in [VERSION.md](VERSION.md)

### Code Patterns
```typescript
// Plugin module pattern
@NgModule({
    providers: [
        { provide: SomeProvider, useClass: MyProvider, multi: true },
    ],
})
export default class MyModule { }

// Service pattern
@Injectable()
export class MyService {
    constructor(private config: ConfigService) {}
}

// Component pattern
@Component({
    selector: 'my-component',
    templateUrl: './my.component.pug',
    styleUrls: ['./my.component.scss'],
})
export class MyComponent extends BaseComponent { }
```

### File Naming
- Components: `foo.component.ts`, `foo.component.pug`, `foo.component.scss`
- Services: `foo.service.ts`
- API: `api.ts` or `api/index.ts`
- Config: `config.ts`
