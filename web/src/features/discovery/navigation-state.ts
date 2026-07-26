export const navigationEntryPoint = {
  discovery: 'discovery',
  property: 'property',
} as const

export function readDiscoveryNavigationState(state: object) {
  const discoveryIndex =
    'discoveryIndex' in state &&
    typeof state.discoveryIndex === 'number' &&
    Number.isInteger(state.discoveryIndex) &&
    state.discoveryIndex >= 0
      ? state.discoveryIndex
      : undefined
  const entryPoint =
    'entryPoint' in state &&
    (state.entryPoint === navigationEntryPoint.discovery ||
      state.entryPoint === navigationEntryPoint.property)
      ? state.entryPoint
      : undefined

  return { discoveryIndex, entryPoint }
}
