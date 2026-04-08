#!/usr/bin/env node

/**
 * bump-version.mjs — Synchronize version across all package.json files
 * 
 * Reads the version from VERSION.md (single source of truth) and updates
 * all package.json files in the monorepo to match.
 * 
 * Usage:
 *   node scripts/bump-version.mjs [patch|minor|major|nightly|<version>]
 * 
 * Examples:
 *   node scripts/bump-version.mjs patch     # 1.0.231-nightly.0 → 1.0.232-nightly.0
 *   node scripts/bump-version.mjs minor     # 1.0.231-nightly.0 → 1.1.231-nightly.0
 *   node scripts/bump-version.mjs major     # 1.0.231-nightly.0 → 2.0.231-nightly.0
 *   node scripts/bump-version.mjs nightly   # Just increments nightly counter
 *   node scripts/bump-version.mjs 1.0.232   # Set to specific version
 *   node scripts/bump-version.mjs           # Just syncs without bumping
 */

import * as fs from 'fs'
import * as path from 'path'
import { fileURLToPath } from 'url'

const __filename = fileURLToPath(import.meta.url)
const __dirname = path.dirname(__filename)
const ROOT = path.resolve(__dirname, '..')

const VERSION_FILE = path.join(ROOT, 'VERSION.md')

/**
 * Read the current version from VERSION.md
 */
function readVersion() {
    const content = fs.readFileSync(VERSION_FILE, 'utf-8').trim()
    // Extract version string - could be just "1.0.231-nightly.0" or have extra content
    const match = content.match(/^(\d+\.\d+\.\d+(?:-[a-zA-Z0-9.]+)?)/m)
    if (!match) {
        throw new Error(`Could not parse version from VERSION.md: ${content}`)
    }
    return match[1]
}

/**
 * Write version to VERSION.md
 */
function writeVersion(version) {
    fs.writeFileSync(VERSION_FILE, version + '\n')
    console.log(`Updated VERSION.md to ${version}`)
}

/**
 * Bump version according to semver rules
 */
function bumpVersion(currentVersion, bumpType) {
    // Parse the version: major.minor.patch[-prerelease]
    const parts = currentVersion.match(/^(\d+)\.(\d+)\.(\d+)(?:-(.+))?$/)
    if (!parts) {
        throw new Error(`Invalid version format: ${currentVersion}`)
    }

    let [, major, minor, patch, prerelease] = parts
    major = parseInt(major)
    minor = parseInt(minor)
    patch = parseInt(patch)

    switch (bumpType) {
        case 'major':
            major++
            minor = 0
            patch = 0
            prerelease = 'nightly.0'
            break
        case 'minor':
            minor++
            patch = 0
            prerelease = 'nightly.0'
            break
        case 'patch':
            patch++
            prerelease = 'nightly.0'
            break
        case 'nightly':
            if (prerelease && prerelease.startsWith('nightly.')) {
                const nightlyNum = parseInt(prerelease.split('.')[1]) + 1
                prerelease = `nightly.${nightlyNum}`
            } else {
                prerelease = 'nightly.0'
            }
            break
        default:
            // Assume it's a specific version string
            return bumpType
    }

    const version = `${major}.${minor}.${patch}`
    return prerelease ? `${version}-${prerelease}` : version
}

/**
 * Find all package.json files that should be updated
 */
function findPackageJsonFiles() {
    const files = []
    const entries = fs.readdirSync(ROOT, { withFileTypes: true })
    
    for (const entry of entries) {
        if (entry.isDirectory() && entry.name.startsWith('tabby-')) {
            const pkgPath = path.join(ROOT, entry.name, 'package.json')
            if (fs.existsSync(pkgPath)) {
                files.push(pkgPath)
            }
        }
    }
    
    // Also update app/package.json
    const appPkgPath = path.join(ROOT, 'app', 'package.json')
    if (fs.existsSync(appPkgPath)) {
        files.push(appPkgPath)
    }
    
    return files
}

/**
 * Update version in a package.json file
 */
function updatePackageVersion(filePath, newVersion) {
    const content = fs.readFileSync(filePath, 'utf-8')
    const pkg = JSON.parse(content)
    
    const oldVersion = pkg.version
    pkg.version = newVersion
    
    fs.writeFileSync(filePath, JSON.stringify(pkg, null, 2) + '\n')
    
    if (oldVersion !== newVersion) {
        console.log(`  Updated ${path.relative(ROOT, filePath)}: ${oldVersion} → ${newVersion}`)
    }
    
    return oldVersion !== newVersion
}

/**
 * Update CHANGELOG.md with new version section
 */
function updateChangelog(newVersion) {
    const changelogPath = path.join(ROOT, 'CHANGELOG.md')
    if (!fs.existsSync(changelogPath)) {
        console.log('  CHANGELOG.md not found, skipping')
        return
    }
    
    const content = fs.readFileSync(changelogPath, 'utf-8')
    const today = new Date().toISOString().split('T')[0]
    
    // Check if this version already has a section
    if (content.includes(`[${newVersion}]`)) {
        console.log(`  CHANGELOG.md already has section for ${newVersion}`)
        return
    }
    
    // Insert new version section after the header
    const insertMarker = '## ['
    const insertIndex = content.indexOf(insertMarker)
    
    if (insertIndex === -1) {
        console.log('  Could not find insertion point in CHANGELOG.md')
        return
    }
    
    const newSection = `## [${newVersion}] - ${today}\n\n### Changed\n- Version bump\n\n`
    
    const newContent = content.slice(0, insertIndex) + newSection + content.slice(insertIndex)
    fs.writeFileSync(changelogPath, newContent)
    console.log(`  Added ${newVersion} section to CHANGELOG.md`)
}

// Main
const args = process.argv.slice(2)
const currentVersion = readVersion()
console.log(`Current version: ${currentVersion}`)

let newVersion

if (args.length > 0) {
    newVersion = bumpVersion(currentVersion, args[0])
    writeVersion(newVersion)
    console.log(`Bumped to: ${newVersion}`)
} else {
    newVersion = currentVersion
    console.log('Syncing version (no bump)...')
}

// Update all package.json files
console.log('\nUpdating package.json files:')
const pkgFiles = findPackageJsonFiles()
let updated = 0
for (const file of pkgFiles) {
    if (updatePackageVersion(file, newVersion)) {
        updated++
    }
}
console.log(`\nUpdated ${updated} of ${pkgFiles.length} package.json files`)

// Update CHANGELOG.md
console.log('\nUpdating CHANGELOG.md:')
if (args.length > 0) {
    updateChangelog(newVersion)
}

console.log('\nDone!')
