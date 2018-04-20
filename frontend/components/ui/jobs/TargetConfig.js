import React from 'react'

import Select from 'react-select'

import {
  Flex, Box, Grid,
  Container,
  Heading
} from 'ooni-components'

const TargetConfig = (props) => {
  return (
    <Flex>
      <Box w={1/2} pr={2}>
      <Heading h={4}>Country</Heading>
      <Select
        name='countries'
        multi
        options={props.countries}
        value={props.targetCountries}
        onChange={props.onTargetCountryChange}
      />
      </Box>

      <Box w={1/2}>
      <Heading h={4}>Platform</Heading>
      <Select
        name='platform'
        multi
        options={props.platforms}
        value={props.targetPlatforms}
        onChange={props.onTargetPlatformChange}
      />
      </Box>
    </Flex>
  )
}

export default TargetConfig
