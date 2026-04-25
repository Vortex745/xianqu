import test from 'node:test'
import assert from 'node:assert/strict'

import { validatePublishImageUpload } from '../publishUploadPolicy.js'

test('allows oversized images because publish uploads are compressed after upload', () => {
  const result = validatePublishImageUpload({
    name: 'huge-photo.jpg',
    size: 500 * 1024 * 1024,
    type: 'image/jpeg'
  })

  assert.equal(result.allowed, true)
  assert.equal(result.message, '')
})
