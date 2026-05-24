#!/usr/bin/env node
import { build as builder } from 'electron-builder'
import * as vars from './vars.mjs'
import { execSync } from 'child_process'

console.info('Building Tabby for Staging...')

// 1. Force production Go build with optimizations
console.info('Compiling optimized Go backend...')
execSync('npm run build:go', { stdio: 'inherit' })

// 2. Run performance benchmark to ensure baseline is met
console.info('Running performance benchmark...')
try {
    execSync('npm run test:integration', { stdio: 'inherit' })
} catch (e) {
    console.error('Performance benchmark failed, aborting staging build.')
    process.exit(1)
}

// 3. Package the app
builder({
    dir: true,
    config: {
        npmRebuild: false,
        extraMetadata: {
            version: vars.version + '-staging',
        },
        // Force certain staging-specific configurations if needed
    },
    publish: 'never',
}).then(() => {
    console.info('Staging build complete. Artifacts in app/dist/')
}).catch(e => {
    console.error('Staging build failed:', e)
    process.exit(1)
})
