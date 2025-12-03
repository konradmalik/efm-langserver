# efm-langserver

- [1. Property `commands`](#commands)
- [2. Property `languages`](#languages)
  - [2.1. Pattern Property `^([a-z0-9_-]+)+$`](#languages_pattern1)
    - [2.1.1. tool-definition](#languages_pattern1_items)
      - [2.1.1.1. Property `prefix`](#languages_pattern1_items_prefix)
      - [2.1.1.2. Property `format-can-range`](#languages_pattern1_items_format-can-range)
      - [2.1.1.3. Property `format-command`](#languages_pattern1_items_format-command)
      - [2.1.1.4. Property `env`](#languages_pattern1_items_env)
        - [2.1.1.4.1. env items](#languages_pattern1_items_env_items)
      - [2.1.1.5. Property `lint-command`](#languages_pattern1_items_lint-command)
      - [2.1.1.6. Property `lint-offset-columns`](#languages_pattern1_items_lint-offset-columns)
      - [2.1.1.7. Property `lint-category-map`](#languages_pattern1_items_lint-category-map)
      - [2.1.1.8. Property `lint-formats`](#languages_pattern1_items_lint-formats)
        - [2.1.1.8.1. lint-formats items](#languages_pattern1_items_lint-formats_items)
      - [2.1.1.9. Property `lint-ignore-exit-code`](#languages_pattern1_items_lint-ignore-exit-code)
      - [2.1.1.10. Property `lint-offset`](#languages_pattern1_items_lint-offset)
      - [2.1.1.11. Property `lint-after-open`](#languages_pattern1_items_lint-after-open)
      - [2.1.1.12. Property `lint-on-save`](#languages_pattern1_items_lint-on-save)
      - [2.1.1.13. Property `lint-on-change`](#languages_pattern1_items_lint-on-change)
      - [2.1.1.14. Property `lint-severity`](#languages_pattern1_items_lint-severity)
      - [2.1.1.15. Property `lint-source`](#languages_pattern1_items_lint-source)
      - [2.1.1.16. Property `lint-stdin`](#languages_pattern1_items_lint-stdin)
      - [2.1.1.17. Property `root-markers`](#languages_pattern1_items_root-markers)
        - [2.1.1.17.1. root-markers items](#languages_pattern1_items_root-markers_items)
      - [2.1.1.18. Property `require-marker`](#languages_pattern1_items_require-marker)
- [3. Property `tools`](#tools)
  - [3.1. Pattern Property `tool-definition`](#tools_pattern1)
- [4. Property `root-markers`](#root-markers)
  - [4.1. root-markers items](#root-markers_items)
- [5. Property `log-level`](#log-level)
- [6. Property `format-debounce`](#format-debounce)
- [7. Property `lint-debounce`](#lint-debounce)

**Title:** efm-langserver

|                           |             |
| ------------------------- | ----------- |
| **Type**                  | `object`    |
| **Required**              | No          |
| **Additional properties** | Not allowed |

**Description:** If configuring via `DidChangeConfiguration` (e.g. an editor API such as `nvim-lspconfig`), all properties should be in camelCase instead of kebab-case.

| Property                               | Pattern | Type            | Deprecated | Definition                          | Title/Description                                               |
| -------------------------------------- | ------- | --------------- | ---------- | ----------------------------------- | --------------------------------------------------------------- |
| - [commands](#commands )               | No      | object          | No         | In #/definitions/command-definition | -                                                               |
| - [languages](#languages )             | No      | object          | No         | -                                   | list of language                                                |
| - [tools](#tools )                     | No      | object          | No         | -                                   | definition of tools                                             |
| - [root-markers](#root-markers )       | No      | array of string | No         | -                                   | markers to find root directory                                  |
| - [log-level](#log-level )             | No      | number          | No         | -                                   | log level                                                       |
| - [format-debounce](#format-debounce ) | No      | string          | No         | -                                   | duration to debounce calls to the formatter executable. e.g: 1s |
| - [lint-debounce](#lint-debounce )     | No      | string          | No         | -                                   | duration to debounce calls to the linter executable. e.g.: 1s   |

## <a name="commands"></a>1. Property `commands`

|                           |                                  |
| ------------------------- | -------------------------------- |
| **Type**                  | `object`                         |
| **Required**              | No                               |
| **Additional properties** | Any type allowed                 |
| **Defined in**            | #/definitions/command-definition |

## <a name="languages"></a>2. Property `languages`

|                           |                  |
| ------------------------- | ---------------- |
| **Type**                  | `object`         |
| **Required**              | No               |
| **Additional properties** | Any type allowed |

**Description:** list of language

| Property                                   | Pattern | Type  | Deprecated | Definition | Title/Description |
| ------------------------------------------ | ------- | ----- | ---------- | ---------- | ----------------- |
| - [^([a-z0-9_-]+)+$](#languages_pattern1 ) | Yes     | array | No         | -          | -                 |

### <a name="languages_pattern1"></a>2.1. Pattern Property `^([a-z0-9_-]+)+$`
> All properties whose name matches the regular expression
```^([a-z0-9_-]+)+$``` ([Test](https://regex101.com/?regex=%5E%28%5Ba-z0-9_-%5D%2B%29%2B%24))
must respect the following conditions

|              |         |
| ------------ | ------- |
| **Type**     | `array` |
| **Required** | No      |

|                      | Array restrictions |
| -------------------- | ------------------ |
| **Min items**        | N/A                |
| **Max items**        | N/A                |
| **Items unicity**    | False              |
| **Additional items** | False              |
| **Tuple validation** | See below          |

| Each item of this array must be              | Description            |
| -------------------------------------------- | ---------------------- |
| [tool-definition](#languages_pattern1_items) | definition of the tool |

#### <a name="languages_pattern1_items"></a>2.1.1. tool-definition

|                           |                               |
| ------------------------- | ----------------------------- |
| **Type**                  | `object`                      |
| **Required**              | No                            |
| **Additional properties** | Not allowed                   |
| **Defined in**            | #/definitions/tool-definition |

**Description:** definition of the tool

| Property                                                                    | Pattern | Type            | Deprecated | Definition | Title/Description                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                       |
| --------------------------------------------------------------------------- | ------- | --------------- | ---------- | ---------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| - [prefix](#languages_pattern1_items_prefix )                               | No      | string          | No         | -          | If `lint-source` doesn't work, you can set a prefix here instead, which will render the messages as "[prefix] message".                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               |
| - [format-can-range](#languages_pattern1_items_format-can-range )           | No      | boolean         | No         | -          | Whether the formatting command handles range start and range end                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        |
| - [format-command](#languages_pattern1_items_format-command )               | No      | string          | No         | -          | Formatting command. Input filename can be injected using `${INPUT}`, and flags can be injected using `${--flag:key}` (adds `--flag <value>` if value exists for key), `${--flag=key}` (adds `--flag=<value>` if value exists for key), or `${--flag:!key}` (adds `--flag` if value for key is falsy).<br /><br />`efm-langserver` may provide values for keys `charStart`, `charEnd`, `rowStart`, `rowEnd`, `colStart`, `colEnd`, or any key in [`interface FormattingOptions`](https://microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/#formattingOptions).<br /><br />Example: `prettier --stdin --stdin-filepath ${INPUT} ${--tab-width:tabWidth} ${--use-tabs:insertSpaces} ${--range-start=charStart} ${--range-start=charEnd}` |
| - [env](#languages_pattern1_items_env )                                     | No      | array of string | No         | -          | command environment variables and values                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                |
| - [lint-command](#languages_pattern1_items_lint-command )                   | No      | string          | No         | -          | Lint command. Input filename can be injected using `${INPUT}`.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        |
| - [lint-offset-columns](#languages_pattern1_items_lint-offset-columns )     | No      | number          | No         | -          | offset value to skip columns (will be added)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                            |
| - [lint-category-map](#languages_pattern1_items_lint-category-map )         | No      | object          | No         | -          | Map linter categories to LSP categories                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 |
| - [lint-formats](#languages_pattern1_items_lint-formats )                   | No      | array of string | No         | -          | List of Vim errorformats to capture. See: https://vimhelp.org/quickfix.txt.html#errorformats. If this is not expressive enough, you can edit the `lint-command` to do some preprocessing, e.g. using `sed` or `jq`.<br /><br />`efm-langserver` uses a Go implementation to parse the errors, which comes with a CLI for quick testing: https://github.com/reviewdog/errorformat                                                                                                                                                                                                                                                                                                                                                                                                                |
| - [lint-ignore-exit-code](#languages_pattern1_items_lint-ignore-exit-code ) | No      | boolean         | No         | -          | ignore exit code of lint                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                |
| - [lint-offset](#languages_pattern1_items_lint-offset )                     | No      | number          | No         | -          | offset value to skip lines (will be subtracted)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                         |
| - [lint-after-open](#languages_pattern1_items_lint-after-open )             | No      | boolean         | No         | -          | lint after open (defaults to true)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      |
| - [lint-on-save](#languages_pattern1_items_lint-on-save )                   | No      | boolean         | No         | -          | lint on save (defaults to true)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                         |
| - [lint-on-change](#languages_pattern1_items_lint-on-change )               | No      | boolean         | No         | -          | lint on change (defaults to true)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                       |
| - [lint-severity](#languages_pattern1_items_lint-severity )                 | No      | number          | No         | -          | default severity to show if violation doesn't provide severity. 1 = error, 2 = warning, 3 = info, 4 = hint                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              |
| - [lint-source](#languages_pattern1_items_lint-source )                     | No      | string          | No         | -          | show where the lint came from, e.g. 'eslint'                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                            |
| - [lint-stdin](#languages_pattern1_items_lint-stdin )                       | No      | boolean         | No         | -          | use stdin for the lint                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                  |
| - [root-markers](#languages_pattern1_items_root-markers )                   | No      | array of string | No         | -          | markers to find root directory                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          |
| - [require-marker](#languages_pattern1_items_require-marker )               | No      | boolean         | No         | -          | require a marker to run linter                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          |

##### <a name="languages_pattern1_items_prefix"></a>2.1.1.1. Property `prefix`

|              |          |
| ------------ | -------- |
| **Type**     | `string` |
| **Required** | No       |

**Description:** If `lint-source` doesn't work, you can set a prefix here instead, which will render the messages as "[prefix] message".

##### <a name="languages_pattern1_items_format-can-range"></a>2.1.1.2. Property `format-can-range`

|              |           |
| ------------ | --------- |
| **Type**     | `boolean` |
| **Required** | No        |

**Description:** Whether the formatting command handles range start and range end

##### <a name="languages_pattern1_items_format-command"></a>2.1.1.3. Property `format-command`

|              |          |
| ------------ | -------- |
| **Type**     | `string` |
| **Required** | No       |

**Description:** Formatting command. Input filename can be injected using `${INPUT}`, and flags can be injected using `${--flag:key}` (adds `--flag <value>` if value exists for key), `${--flag=key}` (adds `--flag=<value>` if value exists for key), or `${--flag:!key}` (adds `--flag` if value for key is falsy).

`efm-langserver` may provide values for keys `charStart`, `charEnd`, `rowStart`, `rowEnd`, `colStart`, `colEnd`, or any key in [`interface FormattingOptions`](https://microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/#formattingOptions).

Example: `prettier --stdin --stdin-filepath ${INPUT} ${--tab-width:tabWidth} ${--use-tabs:insertSpaces} ${--range-start=charStart} ${--range-start=charEnd}`

##### <a name="languages_pattern1_items_env"></a>2.1.1.4. Property `env`

|              |                   |
| ------------ | ----------------- |
| **Type**     | `array of string` |
| **Required** | No                |

**Description:** command environment variables and values

|                      | Array restrictions |
| -------------------- | ------------------ |
| **Min items**        | N/A                |
| **Max items**        | N/A                |
| **Items unicity**    | False              |
| **Additional items** | False              |
| **Tuple validation** | See below          |

| Each item of this array must be                  | Description |
| ------------------------------------------------ | ----------- |
| [env items](#languages_pattern1_items_env_items) | -           |

###### <a name="languages_pattern1_items_env_items"></a>2.1.1.4.1. env items

|              |          |
| ------------ | -------- |
| **Type**     | `string` |
| **Required** | No       |

| Restrictions                      |                                                                     |
| --------------------------------- | ------------------------------------------------------------------- |
| **Must match regular expression** | ```^.+=.+$``` [Test](https://regex101.com/?regex=%5E.%2B%3D.%2B%24) |

##### <a name="languages_pattern1_items_lint-command"></a>2.1.1.5. Property `lint-command`

|              |          |
| ------------ | -------- |
| **Type**     | `string` |
| **Required** | No       |

**Description:** Lint command. Input filename can be injected using `${INPUT}`.

##### <a name="languages_pattern1_items_lint-offset-columns"></a>2.1.1.6. Property `lint-offset-columns`

|              |          |
| ------------ | -------- |
| **Type**     | `number` |
| **Required** | No       |

**Description:** offset value to skip columns (will be added)

##### <a name="languages_pattern1_items_lint-category-map"></a>2.1.1.7. Property `lint-category-map`

|                           |                  |
| ------------------------- | ---------------- |
| **Type**                  | `object`         |
| **Required**              | No               |
| **Additional properties** | Any type allowed |

**Description:** Map linter categories to LSP categories

##### <a name="languages_pattern1_items_lint-formats"></a>2.1.1.8. Property `lint-formats`

|              |                   |
| ------------ | ----------------- |
| **Type**     | `array of string` |
| **Required** | No                |

**Description:** List of Vim errorformats to capture. See: https://vimhelp.org/quickfix.txt.html#errorformats. If this is not expressive enough, you can edit the `lint-command` to do some preprocessing, e.g. using `sed` or `jq`.

`efm-langserver` uses a Go implementation to parse the errors, which comes with a CLI for quick testing: https://github.com/reviewdog/errorformat

|                      | Array restrictions |
| -------------------- | ------------------ |
| **Min items**        | N/A                |
| **Max items**        | N/A                |
| **Items unicity**    | False              |
| **Additional items** | False              |
| **Tuple validation** | See below          |

| Each item of this array must be                                    | Description |
| ------------------------------------------------------------------ | ----------- |
| [lint-formats items](#languages_pattern1_items_lint-formats_items) | -           |

###### <a name="languages_pattern1_items_lint-formats_items"></a>2.1.1.8.1. lint-formats items

|              |          |
| ------------ | -------- |
| **Type**     | `string` |
| **Required** | No       |

##### <a name="languages_pattern1_items_lint-ignore-exit-code"></a>2.1.1.9. Property `lint-ignore-exit-code`

|              |           |
| ------------ | --------- |
| **Type**     | `boolean` |
| **Required** | No        |
| **Default**  | `true`    |

**Description:** ignore exit code of lint

##### <a name="languages_pattern1_items_lint-offset"></a>2.1.1.10. Property `lint-offset`

|              |          |
| ------------ | -------- |
| **Type**     | `number` |
| **Required** | No       |

**Description:** offset value to skip lines (will be subtracted)

##### <a name="languages_pattern1_items_lint-after-open"></a>2.1.1.11. Property `lint-after-open`

|              |           |
| ------------ | --------- |
| **Type**     | `boolean` |
| **Required** | No        |
| **Default**  | `true`    |

**Description:** lint after open (defaults to true)

##### <a name="languages_pattern1_items_lint-on-save"></a>2.1.1.12. Property `lint-on-save`

|              |           |
| ------------ | --------- |
| **Type**     | `boolean` |
| **Required** | No        |

**Description:** lint on save (defaults to true)

##### <a name="languages_pattern1_items_lint-on-change"></a>2.1.1.13. Property `lint-on-change`

|              |           |
| ------------ | --------- |
| **Type**     | `boolean` |
| **Required** | No        |

**Description:** lint on change (defaults to true)

##### <a name="languages_pattern1_items_lint-severity"></a>2.1.1.14. Property `lint-severity`

|              |          |
| ------------ | -------- |
| **Type**     | `number` |
| **Required** | No       |

**Description:** default severity to show if violation doesn't provide severity. 1 = error, 2 = warning, 3 = info, 4 = hint

##### <a name="languages_pattern1_items_lint-source"></a>2.1.1.15. Property `lint-source`

|              |          |
| ------------ | -------- |
| **Type**     | `string` |
| **Required** | No       |

**Description:** show where the lint came from, e.g. 'eslint'

##### <a name="languages_pattern1_items_lint-stdin"></a>2.1.1.16. Property `lint-stdin`

|              |           |
| ------------ | --------- |
| **Type**     | `boolean` |
| **Required** | No        |
| **Default**  | `true`    |

**Description:** use stdin for the lint

##### <a name="languages_pattern1_items_root-markers"></a>2.1.1.17. Property `root-markers`

|              |                   |
| ------------ | ----------------- |
| **Type**     | `array of string` |
| **Required** | No                |

**Description:** markers to find root directory

|                      | Array restrictions |
| -------------------- | ------------------ |
| **Min items**        | N/A                |
| **Max items**        | N/A                |
| **Items unicity**    | False              |
| **Additional items** | False              |
| **Tuple validation** | See below          |

| Each item of this array must be                                    | Description |
| ------------------------------------------------------------------ | ----------- |
| [root-markers items](#languages_pattern1_items_root-markers_items) | -           |

###### <a name="languages_pattern1_items_root-markers_items"></a>2.1.1.17.1. root-markers items

|              |          |
| ------------ | -------- |
| **Type**     | `string` |
| **Required** | No       |

##### <a name="languages_pattern1_items_require-marker"></a>2.1.1.18. Property `require-marker`

|              |           |
| ------------ | --------- |
| **Type**     | `boolean` |
| **Required** | No        |

**Description:** require a marker to run linter

## <a name="tools"></a>3. Property `tools`

|                           |                  |
| ------------------------- | ---------------- |
| **Type**                  | `object`         |
| **Required**              | No               |
| **Additional properties** | Any type allowed |

**Description:** definition of tools

| Property                               | Pattern | Type   | Deprecated | Definition                                                     | Title/Description      |
| -------------------------------------- | ------- | ------ | ---------- | -------------------------------------------------------------- | ---------------------- |
| - [^([a-z0-9_-]+)+$](#tools_pattern1 ) | Yes     | object | No         | Same as [languages_pattern1_items](#languages_pattern1_items ) | definition of the tool |

### <a name="tools_pattern1"></a>3.1. Pattern Property `tool-definition`
> All properties whose name matches the regular expression
```^([a-z0-9_-]+)+$``` ([Test](https://regex101.com/?regex=%5E%28%5Ba-z0-9_-%5D%2B%29%2B%24))
must respect the following conditions

|                           |                                                       |
| ------------------------- | ----------------------------------------------------- |
| **Type**                  | `object`                                              |
| **Required**              | No                                                    |
| **Additional properties** | Not allowed                                           |
| **Same definition as**    | [languages_pattern1_items](#languages_pattern1_items) |

**Description:** definition of the tool

## <a name="root-markers"></a>4. Property `root-markers`

|              |                   |
| ------------ | ----------------- |
| **Type**     | `array of string` |
| **Required** | No                |

**Description:** markers to find root directory

|                      | Array restrictions |
| -------------------- | ------------------ |
| **Min items**        | N/A                |
| **Max items**        | N/A                |
| **Items unicity**    | False              |
| **Additional items** | False              |
| **Tuple validation** | See below          |

| Each item of this array must be           | Description |
| ----------------------------------------- | ----------- |
| [root-markers items](#root-markers_items) | -           |

### <a name="root-markers_items"></a>4.1. root-markers items

|              |          |
| ------------ | -------- |
| **Type**     | `string` |
| **Required** | No       |

## <a name="log-level"></a>5. Property `log-level`

|              |          |
| ------------ | -------- |
| **Type**     | `number` |
| **Required** | No       |

**Description:** log level

| Restrictions |        |
| ------------ | ------ |
| **Minimum**  | &ge; 1 |

## <a name="format-debounce"></a>6. Property `format-debounce`

|              |          |
| ------------ | -------- |
| **Type**     | `string` |
| **Required** | No       |

**Description:** duration to debounce calls to the formatter executable. e.g: 1s

## <a name="lint-debounce"></a>7. Property `lint-debounce`

|              |          |
| ------------ | -------- |
| **Type**     | `string` |
| **Required** | No       |

**Description:** duration to debounce calls to the linter executable. e.g.: 1s

----------------------------------------------------------------------------------------------------------------------------
Generated using [json-schema-for-humans](https://github.com/coveooss/json-schema-for-humans) on 2025-12-03 at 10:05:58 +0100
