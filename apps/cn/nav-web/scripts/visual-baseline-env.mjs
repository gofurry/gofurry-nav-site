// This guards baseline updates only; local comparison remains diagnostic.
if (process.platform !== 'linux'
  || Number(process.versions.node.split('.')[0]) !== 24
  || process.env.GOFURRY_VISUAL_ENV !== 'pinned') {
  console.error('Visual baselines may only be updated in the pinned GoFurry visual environment (Linux, Node 24, GOFURRY_VISUAL_ENV=pinned).')
  process.exit(1)
}
