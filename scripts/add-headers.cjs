const fs = require('fs');
const path = require('path');
const { promisify } = require('util');
const readFile = promisify(fs.readFile);
const writeFile = promisify(fs.writeFile);

const copyright = `/*
 * Copyright (c) 2023-2025 KCai Technology (https://kcaitech.com). All rights reserved.
 *
 * This file is part of the Vextra project, which is licensed under the AGPL-3.0 license.
 * The full license text can be found in the LICENSE file in the root directory of this source tree.
 *
 * For more information about the AGPL-3.0 license, please visit:
 * https://www.gnu.org/licenses/agpl-3.0.html
 */

`;

async function addHeaderToFile(filePath) {
    try {
        const content = await readFile(filePath, 'utf8');
        if (!content.includes(copyright)) {
            await writeFile(filePath, copyright + content);
            console.log(`✅ Added header to ${filePath}`);
        } else {
            console.log(`⏭️  Skipped ${filePath} (header already exists)`);
        }
    } catch (err) {
        console.error(`❌ Error processing ${filePath}:`, err);
    }
}

const includes = ['.']
const excludes = ['assets', 'dist', 'node_modules', 'db', 'scripts']

function isExclude(filePath) {
    for (let i = 0, len = excludes.length; i < len; ++i) {
        if (filePath.endsWith(excludes[i])) return true
    }
    return false
}

function isCodeFile(filePath) {
    return filePath.endsWith('.ts') || filePath.endsWith('.vue') || filePath.endsWith('.go')
}

function findTypeScriptFiles(dir) {
    if (!fs.existsSync(dir)) return
    const files = fs.readdirSync(dir);
    
    files.forEach(file => {
        const filePath = path.join(dir, file);
        const stat = fs.statSync(filePath);
        
        if (stat.isDirectory()) {
            if (!isExclude(filePath)) findTypeScriptFiles(filePath);
        } else if (isCodeFile(filePath)) {
            addHeaderToFile(filePath);
        }
    });
}

for (let i = 0, len = includes.length; i < len; ++i) {
    findTypeScriptFiles(path.resolve(__dirname, '..', includes[i]));
}