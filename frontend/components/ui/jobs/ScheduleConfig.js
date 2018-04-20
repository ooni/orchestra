import React from 'react'

import Head from 'next/head'

import Button from 'material-ui/Button'
import Checkbox from 'material-ui/Checkbox'
import { FormGroup, FormControlLabel } from 'material-ui/Form'
import Input, { InputLabel } from 'material-ui/Input'

import MomentUtils from 'material-ui-pickers/utils/moment-utils'
import MuiPickersUtilsProvider from 'material-ui-pickers/utils/MuiPickersUtilsProvider'
import DateTimePicker from 'material-ui-pickers/DateTimePicker'

import {
  DurationPicker,
  RepeatString,
} from '../schedule'

import {
  Flex, Box, Grid,
  Container,
  Heading
} from 'ooni-components'

import moment from 'moment'

const ScheduleConfig = (props) => {
  const {
    startMoment,
    onStartMomentChange,
    repeatCount,
    onRepeatChange,
    duration,
    onDurationChange
  } = props

  return (
    <MuiPickersUtilsProvider utils={MomentUtils}>
    <Flex>
    <Box px={2}>
      <div className='option'>

        <DateTimePicker
          value={startMoment}
          disablePast
          label='Start on'
          onChange={onStartMomentChange}
        />
        <Button onClick={() => onStartMomentChange(moment(new Date()))}>Now</Button>
      </div>
    </Box>

    <Box px={2}>
      <FormGroup>
      <FormControlLabel
        control={
          <Checkbox
          checked={repeatCount !== 1}
          onChange={({target}) => {
            if (target.checked === true) onRepeatChange(2)
            else onRepeatChange(1)
          }}
          />
        }
        label="Repeat"
        />
      {repeatCount !== 1 && <div>
        <FormControlLabel
        control={
          <Checkbox
          checked={repeatCount === 0}
          onChange={({target}) => {
            if (target.checked === true) onRepeatChange(0)
            else onRepeatChange(2)
          }}
        />
        }
        label="Repeat forever"
        />

        {repeatCount !== 0 && <Input
        type='number'
        min='0'
        placeholder='times to repeat'
        name='repeat-count'
        value={repeatCount}
        onChange={({target}) => {onRepeatChange(target.value)}}
        />}
      <DurationPicker onChange={onDurationChange}
                      duration={duration} />

      <RepeatString duration={duration} repeatCount={repeatCount} />
      </div>}
      </FormGroup>
    </Box>
    </Flex>
    </MuiPickersUtilsProvider>
  )
}

export default ScheduleConfig
