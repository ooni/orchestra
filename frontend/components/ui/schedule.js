import React from 'react'
import PropTypes from 'prop-types'

import Immutable from 'immutable'

import moment from 'moment'

import AlarmIcon from '@material-ui/icons/Alarm'
//import Slider from 'react-toolbox/lib/slider/Slider'

import { Flex, Box } from 'ooni-components'

const Slider = () => <div>Not implemented</div>

export class DesignatorSlider extends React.Component {
  static propTypes = {
    unit: PropTypes.string.isRequired,
    min: PropTypes.number,
    max: PropTypes.number,
    step: PropTypes.number,
    defaultValue: PropTypes.number,
    onChange: PropTypes.func
  }

  static defaultProps = {
    min: 0,
    max: 60,
    step: 1,
    defaultValue: 0
  }

  constructor(props) {
    super(props)
      console.log("Setting value to", props.defaultValue)
    this.state = {
      value: props.defaultValue
    }
    this.onChange = this.onChange.bind(this)
  }

  onChange (value) {
    this.setState({value})
    this.props.onChange(this.props.unit, value)
  }

  render () {
    return(
      <Slider
        editable
        pinned snaps
        min={this.props.min}
        max={this.props.max}
        step={this.props.step}
        value={this.state.value}
        onChange={this.onChange}
      />
    )
  }

}

export class DurationPicker extends React.Component {
  static propTypes = {
    onChange: PropTypes.func.isRequired,
    duration: PropTypes.object
  }

  constructor(props) {
    super(props)
    this.state = {
      duration: Immutable.Map({
        Y: props.duration.Y || 0,
        M: props.duration.M || 0,
        W: props.duration.W || 0,
        D: props.duration.D || 0,
        h: props.duration.h || 0,
        m: props.duration.m || 0,
        s: props.duration.s || 0
      }),
      repeat: 0,
      isOpen: false
    }
    this.setDuration = this.setDuration.bind(this);
  }

  setDuration (unit, quantity) {
    let duration = this.state.duration.set(unit, quantity)
    this.setState({duration})
    this.props.onChange(duration.toObject())
  }

  render () {
    const {
      duration
    } = this.state
    return (
      <div className='picker' style={{width: '400px'}}>
        <p>months</p>
        <DesignatorSlider defaultValue={duration.get("M")} unit="M" max={12} onChange={this.setDuration} />
        <p>weeks</p>
        <DesignatorSlider defaultValue={duration.get("W")} unit="W" max={4} onChange={this.setDuration} />
        <p>days</p>
        <DesignatorSlider defaultValue={duration.get("D")} unit="D" max={30} onChange={this.setDuration} />
        <p>hours</p>
        <DesignatorSlider defaultValue={duration.get("h")} unit="h" max={24} onChange={this.setDuration} />
        <p>minutes</p>
        <DesignatorSlider defaultValue={duration.get("m")} unit="m" max={60} onChange={this.setDuration} />
        <p>seconds</p>
        <DesignatorSlider defaultValue={duration.get("s")} unit="s" max={60} onChange={this.setDuration} />
      </div>
    )
  }
}

export const RepeatString = ({duration, repeatCount}) => {
  let units = [
    {'key': 'Y', 'name': 'year'},
    {'key': 'M', 'name': 'month'},
    {'key': 'W', 'name': 'week'},
    {'key': 'D', 'name': 'day'},
    {'key': 'h', 'name': 'hour'},
    {'key': 'm', 'name': 'minute'},
    {'key': 's', 'name': 'second'}
  ]
  return (
    <Flex pt={2} align='center'>
      <Box><AlarmIcon/></Box>
      <Box pl={2}>
      <div>
      Will run
      {repeatCount === 0
      && <strong> forever</strong> }
      {repeatCount === 1
      && <strong>once</strong>}
      {repeatCount > 1
      && <span> <strong>{repeatCount}</strong> times </span> }
      {repeatCount !== 1 && repeatCount !== 0 && <span> every </span>}
      {repeatCount !== 1 && repeatCount !== 0 && units.map((unit) => {
        const value = duration[unit.key]
        if (value && value !== 0) {
          let unitName = unit.name
          if (value > 1) {
            unitName += 's'
          }
          return <span key={unit.key}><strong>{value} {unitName}</strong> </span>
        }
      })}
      </div>
      </Box>
    </Flex>
  )
}

export const ToScheduleString = ({duration, startMoment, repeatCount}) => {
  let mDuration,
    scheduleString = 'R'

  if (repeatCount > 0) {
    scheduleString += repeatCount
  }
  scheduleString += '/'
  scheduleString += startMoment.toISOString()
  scheduleString += '/'
  mDuration = moment.duration({
    seconds: duration.s,
    minutes: duration.m,
    hours: duration.h,
    days: duration.D,
    weeks: duration.W,
    months: duration.M,
    years: duration.Y
  })
  scheduleString += mDuration.toISOString()
  return scheduleString
}
