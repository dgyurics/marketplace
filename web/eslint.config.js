// eslint.config.js
import js from '@eslint/js';
import vue from 'eslint-plugin-vue';
import tseslint from '@typescript-eslint/eslint-plugin';
import tsparser from '@typescript-eslint/parser';
import vueParser from 'vue-eslint-parser';
import prettier from 'eslint-plugin-prettier';
import importPlugin from 'eslint-plugin-import';

export default [
  js.configs.recommended,
  ...vue.configs['flat/recommended'],

  {
    files: ['**/*.ts'],
    languageOptions: {
      parser: tsparser,
      parserOptions: {
        ecmaVersion: 'latest',
        sourceType: 'module',
        project: './tsconfig.app.json',
        tsconfigRootDir: import.meta.dirname,
      },
      globals: {
        console: 'readonly',
        localStorage: 'readonly',
        window: 'readonly',
        document: 'readonly',
        navigator: 'readonly',
        setTimeout: 'readonly',
        alert: 'readonly',
        IntersectionObserver: 'readonly',
        File: 'readonly',
        HTMLInputElement: 'readonly',
        HTMLElement: 'readonly',
        Event: 'readonly',
        FormData: 'readonly',
        Blob: 'readonly',
      },
    },
    plugins: {
      '@typescript-eslint': tseslint,
      prettier,
      import: importPlugin,
    },
    rules: {
      // Enforce consistent quotes
      'quotes': ['error', 'single'],

      // Variable declaration rules
      'prefer-const': 'error',           // Flag let when const could be used
      'no-var': 'error',                 // No var declarations
      'no-unused-vars': 'off',           // Turn off base rule
      '@typescript-eslint/no-unused-vars': ['error', {
        argsIgnorePattern: '^_',
        varsIgnorePattern: '^_',
        ignoreRestSiblings: true
      }],

      // Code quality rules
      'no-unreachable': 'error',         // Dead code after return/throw
      'no-constant-condition': 'error',  // if (true) or while (true)
      'no-duplicate-case': 'error',      // Duplicate case in switch
      'no-empty': 'error',               // Empty blocks
      'no-extra-boolean-cast': 'error',  // !!Boolean(foo)
      'no-implicit-coercion': 'error',   // Prefer explicit type conversion
      'no-lonely-if': 'error',           // if as the only statement in else
      'no-unneeded-ternary': 'error',    // foo ? true : false
      'no-useless-return': 'error',      // return; at end of function
      'no-else-return': 'error',         // Unnecessary else after return
      'no-useless-concat': 'error',      // "a" + "b" instead of "ab"
      'no-useless-computed-key': 'error', // { ["a"]: 1 } instead of { a: 1 }
      'no-useless-escape': 'error',        // Unnecessary escape characters
      'no-useless-catch': 'error',         // Catch that just re-throws
      'prefer-template': 'error',          // Template literals vs concatenation
      'object-shorthand': 'error',         // { foo: foo } -> { foo }
      'no-nested-ternary': 'warn',         // Nested ? : operators          

      // Import sorting
      'import/order': [
        'error',
        {
          'newlines-between': 'always',
          alphabetize: { order: 'asc', caseInsensitive: true },
          groups: [['builtin', 'external'], 'internal', ['parent', 'sibling', 'index']],
        },
      ],

      // TypeScript
      '@typescript-eslint/explicit-function-return-type': 'off',
      '@typescript-eslint/no-explicit-any': 'warn',
      '@typescript-eslint/prefer-optional-chain': 'error',     // foo && foo.bar -> foo?.bar
      '@typescript-eslint/prefer-nullish-coalescing': 'error', // || -> ??
      '@typescript-eslint/no-unnecessary-condition': 'warn',   // Always true/false conditions
      '@typescript-eslint/no-unnecessary-type-assertion': 'error', // as string when already string

      // Error handling
      '@typescript-eslint/no-unsafe-assignment': 'off',
      '@typescript-eslint/no-unsafe-member-access': 'off',
      '@typescript-eslint/no-unsafe-call': 'off',

      // Hygiene
      'no-console': 'warn',
      'no-debugger': 'warn',
      'prettier/prettier': 'error',
    },
  },

  // Vue files configuration
  {
    files: ['**/*.vue'],
    languageOptions: {
      parser: vueParser,
      parserOptions: {
        parser: tsparser,
        ecmaVersion: 'latest',
        sourceType: 'module',
        extraFileExtensions: ['.vue'],
      },
      globals: {
        console: 'readonly',
        localStorage: 'readonly',
        window: 'readonly',
        document: 'readonly',
        navigator: 'readonly',
        setTimeout: 'readonly',
        alert: 'readonly',
        IntersectionObserver: 'readonly',
        File: 'readonly',
        HTMLInputElement: 'readonly',
        HTMLElement: 'readonly',
        Event: 'readonly',
        FormData: 'readonly',
        Blob: 'readonly',
      },
    },
    plugins: {
      '@typescript-eslint': tseslint,
      vue,
      prettier,
      import: importPlugin,
    },
    rules: {
      // Vue - using correct rule names  
      'vue/block-order': ['error', { order: ['template', 'script', 'style'] }],
      'vue/component-name-in-template-casing': ['error', 'PascalCase'],
      'vue/no-unused-vars': 'error',
      'vue/multi-word-component-names': 'off',
      'vue/no-v-html': 'warn',
      'vue/html-indent': 'off', // Let Prettier handle this
      'vue/max-attributes-per-line': 'off', // Let Prettier handle this
      'vue/html-closing-bracket-newline': 'off', // Let Prettier handle this
      'vue/html-self-closing': 'off', // Personal preference
      'vue/singleline-html-element-content-newline': 'off',
      'vue/multiline-html-element-content-newline': 'off',
      'vue/first-attribute-linebreak': 'off',
      'vue/html-closing-bracket-spacing': 'off',
      'vue/attributes-order': 'warn', // Just warn, don't error      

      // Variable declaration rules (same as TS)
      'prefer-const': 'error',
      'no-var': 'error',
      'no-unused-vars': 'off',
      '@typescript-eslint/no-unused-vars': ['error', {
        argsIgnorePattern: '^_',
        varsIgnorePattern: '^_',
        ignoreRestSiblings: true
      }],

      // Code quality rules (same as TS)
      'no-unreachable': 'error',
      'no-constant-condition': 'error',
      'no-duplicate-case': 'error',
      'no-empty': 'error',
      'no-extra-boolean-cast': 'error',
      'no-implicit-coercion': 'error',
      'no-lonely-if': 'error',
      'no-unneeded-ternary': 'error',
      'no-useless-return': 'error',
      'no-else-return': 'error',
      'no-useless-concat': 'error',
      'no-useless-computed-key': 'error',
      'no-useless-escape': 'error',        // Unnecessary escape characters
      'no-useless-catch': 'error',         // Catch that just re-throws
      'prefer-template': 'error',          // Template literals vs concatenation
      'object-shorthand': 'error',         // { foo: foo } -> { foo }
      'no-nested-ternary': 'warn',         // Nested ? : operators      

      // Import sorting
      'import/order': [
        'error',
        {
          'newlines-between': 'always',
          alphabetize: { order: 'asc', caseInsensitive: true },
          groups: [['builtin', 'external'], 'internal', ['parent', 'sibling', 'index']],
        },
      ],

      // TypeScript - non-type-aware rules only
      '@typescript-eslint/explicit-function-return-type': 'off',
      '@typescript-eslint/no-explicit-any': 'warn',

      // Error handling - relaxed for Vue components
      '@typescript-eslint/no-unsafe-assignment': 'off',
      '@typescript-eslint/no-unsafe-member-access': 'off',
      '@typescript-eslint/no-unsafe-call': 'off',

      // Hygiene
      'no-console': 'warn',
      'no-debugger': 'warn',
      'prettier/prettier': 'error',
    },
  },

  // Config files
  {
    files: ['*.config.ts', '*.config.js'],
    languageOptions: {
      parser: tsparser,
      parserOptions: {
        ecmaVersion: 'latest',
        sourceType: 'module',
      },
      globals: {
        console: 'readonly',
        process: 'readonly',
      },
    },
    plugins: {
      '@typescript-eslint': tseslint,
      prettier,
    },
    rules: {
      '@typescript-eslint/no-explicit-any': 'off',
      'prettier/prettier': 'error',
    },
  },

  // Global ignores
  {
    ignores: ['dist/', 'node_modules/', 'eslint.config.js'],
  },
]