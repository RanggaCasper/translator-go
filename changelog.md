# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.6] - 2026-04-21

### Fixed
- Preserve speaker tags wrapped by square brackets, such as [QIFREY], during Indonesian subtitle post-processing.
- Prevent accidental removal of stage-direction prefixes like - [Coco terengah-engah] in cleaned subtitle lines.
- Improve Indonesian cue line splitting so mixed dialogue and stage direction text is rendered as:
	dialogue line first, then stage-direction line.

### Added
- Regression test coverage for bracketed speaker preservation and stage-direction suffix line splitting.

## [1.0.5] - 2026-04-20

### Improved
- Improve Indonesian subtitle translation detection and cue wrapping.
- Improve VTT cue handling by combining multi-line cue text before translation for better translation quality.
- Apply Indonesian subtitle wrapping heuristics with a two-line output cap to keep cues readable.
- Expand translator test coverage for cue processing and Indonesian post-processing behavior.

## [1.0.4] - 2026-04-18

### Improved
- Optimize subtitle output translation

## [1.0.3] - 2026-04-17

### Improved
- Improve subtitle normalization and tag preservation

## [1.0.2] - 2026-03-11

### Feature
- Remove some informal words from the translation to make it more neutral and suitable for a wider audience.

## [1.0.1] - 2026-02-18

### Fixed
- Fix Startup.sh & Makefile

## [1.0.0] - 2026-02-18

### Added
- Initial public release of Transaltor with Go + Fiber architecture.
- API versioning with base path `POST/GET/PUT/DELETE /api/v1/subtitles`.
- `POST /api/v1/subtitles/translate` for subtitle URL translation (`vtt` and `ass` input).
- `POST /api/v1/subtitles/translate/text` for single sentence/text translation.
- `POST /api/v1/subtitles/translate/batch` for large structured payloads (`data.title` and `data.content[].text`).
- `GET /api/v1/subtitles` with pagination and optional `target_lang` filter.
- `GET /api/v1/subtitles/:id`, `PUT /api/v1/subtitles/:id`, `DELETE /api/v1/subtitles/:id`.
- MySQL for subtitle metadata (`subtitle_id`, language, format, path, size, timestamps).
- File system storage for translated content in `storage/subtitles/*.vtt`.
- Public static file serving via `/storage/subtitles/<filename>.vtt`.
- GORM database integration with auto-migration, soft delete support, and indexed metadata fields.
- VTT parser with timestamp normalization and clean text extraction.
- ASS dialogue parser and conversion pipeline to VTT output.
- Batched translation processing with concurrent chunk execution.
- Fallback handling for failed batches and long-text segmentation.
- Reusable HTTP client with connection pooling for translation requests.
- Output cleanup by removing `<font ...>` tags from translated subtitle content.
- Normalized slash format for stored `file_path` values (`storage/subtitles/...`).
- Indonesian informal style conversion layer for anime-friendly translation tone.
- Standard API response schema for success, error, and paginated responses.
- Operational tooling via Makefile commands for build/run/dev/test/lint/format.
- Dockerfile and `docker-compose.yml` for app + MySQL deployment.
- Environment-based configuration via `.env`.
- GitHub Actions release workflow for cross-platform build and asset publishing.
- Comprehensive `README.md` and Postman collection (`Subtitle_Translator_API.postman_collection.json`) for API usage.

[1.0.0]: https://github.com/RanggaCasper/translator-go/releases/tag/v1.0.0
[1.0.1]: https://github.com/RanggaCasper/translator-go/releases/tag/v1.0.1
[1.0.2]: https://github.com/RanggaCasper/translator-go/releases/tag/v1.0.2
[1.0.3]: https://github.com/RanggaCasper/translator-go/releases/tag/v1.0.3
[1.0.4]: https://github.com/RanggaCasper/translator-go/releases/tag/v1.0.4 
[1.0.5]: https://github.com/RanggaCasper/translator-go/releases/tag/v1.0.5
[1.0.6]: https://github.com/RanggaCasper/translator-go/compare/v1.0.5...main