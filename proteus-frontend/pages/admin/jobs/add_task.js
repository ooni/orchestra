import React from 'react'
import Head from 'next/head'
import Router from 'next/router'

import Select from 'react-select'

import Immutable from 'immutable'

import RaisedButton from 'material-ui/RaisedButton'
import Checkbox from 'material-ui/Checkbox'
import DatePicker from 'material-ui/DatePicker'
import TimePicker from 'material-ui/TimePicker'
import Slider from 'material-ui/Slider'
import TextField from 'material-ui/TextField'
import {Card, CardActions, CardTitle, CardText} from 'material-ui/Card'

import moment from 'moment'

import Layout from '../../../components/layout'
import Session from '../../../components/session'

import { Flex, Box, Grid } from 'reflexbox'

class DesignatorSlider extends React.Component {
  static propTypes = {
    unit: React.PropTypes.string.isRequired,
    min: React.PropTypes.number,
    max: React.PropTypes.number,
    step: React.PropTypes.number,
    defaultValue: React.PropTypes.number,
    onChange: React.PropTypes.func
  }

  static defaultProps = {
    min: 0,
    max: 60,
    step: 1,
    defaultValue: 0
  }

  constructor(props) {
    super(props)
    this.state = {
      value: 0
    }
    this.onChange = this.onChange.bind(this)
  }

  onChange (event, value) {
    this.setState({ value: value })
    this.props.onChange(this.props.unit, value)
  }

  render () {
    return(
      <div>
        <div>
          {this.props.unit}
        </div>
        <Slider
          axis="y"
          style={{height: 100}}
          min={this.props.min}
          max={this.props.max}
          step={this.props.step}
          value={this.state.value}
          defaultValue={this.props.defaultValue}
          onChange={this.onChange}
        />
        <TextField
          name={`unit-${this.props.unit}`}
          style={{width: 20}}
          value={this.state.value}
          onChange={this.onChange}
        />
      </div>
    )
  }

}

class DurationPicker extends React.Component {
  static propTypes = {
    onChange: React.PropTypes.func.isRequired
  }

  constructor(props) {
    super(props)
    this.state = {
      duration: Immutable.Map({
        Y: 0,
        M: 0,
        W: 0,
        D: 0,
        h: 0,
        m: 0,
        s: 0
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
    return (
      <div className='picker'>
        <DesignatorSlider unit="Y" max={10} onChange={this.setDuration} />
        <DesignatorSlider unit="M" max={12} onChange={this.setDuration} />
        <DesignatorSlider unit="W" max={4} onChange={this.setDuration} />
        <DesignatorSlider unit="D" max={30} onChange={this.setDuration} />
        <DesignatorSlider unit="h" max={60} onChange={this.setDuration} />
        <DesignatorSlider unit="m" max={60} onChange={this.setDuration} />
        <DesignatorSlider unit="s" max={60} onChange={this.setDuration} />
        <style jsx>{`
        .picker {
          display: flex;
        }
        .picker > :global(div) {
          padding-right: 20px;
        }
        `}</style>
      </div>
    )
  }
}

const RepeatString = ({duration, repeatCount}) => {
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
    <div>
      Will run
      {repeatCount === 0
      && ' forever every '}
      {repeatCount === 1
      && ' once'}
      {repeatCount > 1
      && ` ${repeatCount} times every `}
      {repeatCount !== 1 && units.map((unit) => {
        const value = duration[unit.key]
        if (value && value !== 0) {
          let unitName = unit.name
          if (value > 1) {
            unitName += 's'
          }
          return <span key={unit.key}>{value} {unitName} </span>
        }
      })}
    </div>
  )
}

const ToScheduleString = ({duration, startMoment, repeatCount}) => {
  let mDuration,
    scheduleString = 'R'

  if (repeatCount > 0) {
    scheduleString += repeatCount
  }
  scheduleString += '/'
  scheduleString += startMoment.toISOString()
  scheduleString += '/'
  console.log("converting", duration)
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

class JobCreateConfirm extends React.Component {
  static propTypes = {
    startMoment: React.PropTypes.object,
    duration: React.PropTypes.object,
    repeatCount: React.PropTypes.number,
    globalCategories: React.PropTypes.array,
    countryCategories: React.PropTypes.array,
    selectedTest: React.PropTypes.object,
    targetCountries: React.PropTypes.array,
    targetPlatforms: React.PropTypes.array,
    urls: React.PropTypes.string,
    comment: React.PropTypes.string
  }

  constructor(props) {
    super(props)
  }

  render() {
    const {
      startMoment,
      duration,
      repeatCount,
      globalCategories,
      countryCategories,
      selectedTest,
      targetCountries,
      targetPlatforms,
      urls,
      comment
    } = this.props

    return (
      <div>
        <CardTitle title="Periodic job summary" />
        <CardText>

        <h3>Job comment</h3>
        <p>{comment.toString()}</p>

        <h3>Start time</h3>
        <p>{startMoment.calendar()} ({startMoment.toString()})</p>

        <h3>Duration</h3>
        <div>{RepeatString({duration, repeatCount})} ({ToScheduleString({
                      duration: duration,
                      startMoment: startMoment,
                      repeatCount: repeatCount
                  })})</div>

        <h3>globalCategories</h3>
        <ul>
        {globalCategories.map((category, key) => {
          return (
            <li key={key}>{category.label} ({category.value})</li>
          )
        })}
        </ul>

        <h3>country categories</h3>
        <ul>
        {countryCategories.map((category, key) => {
          return (
            <li key={key}>{category.label} ({category.value})</li>
          )
        })}
        </ul>

        <h3>selectedTest</h3>
        <p>{selectedTest.label}</p>

        <h3>targetCountries</h3>
        <ul>
        {targetCountries.map((country, key) => {
          return (
            <li key={key}>{country.label} ({country.value})</li>
          )
        })}
        </ul>

        <h3>targetPlatforms</h3>
        <ul>
        {targetPlatforms.map((platform, key) => {
          return (
            <li key={key}>{platform.label}</li>
          )
        })}
        </ul>

        <h3>urls</h3>
        <p>{urls.toString()}</p>
        </CardText>

        <style jsx>{`
        h2, h3, p, ul, div {
          margin-bottom: 16px;
        }
        `}</style>

      </div>
    )
  }
}

export default class AdminJobsAdd extends React.Component {

  constructor (props) {
    super(props)
    this.state = {
      startDate: null,
      startTime: null,
      startMoment: moment(),
      repeatCount: 0,
      globalCategories: [],
      countryCategories: [],
      selectedTest: {},
      targetCountries: [],
      targetPlatforms: [],
      duration: {},
      urls: '',
      inputSelectorOpen: false,
      submitted: false,
      comment: '',
      session: new Session(),
      finalized: null
    }

    this.onCountryCategoryChange = this.onCountryCategoryChange.bind(this)
    this.onGlobalCategoryChange = this.onGlobalCategoryChange.bind(this)
    this.onTestChange = this.onTestChange.bind(this)
    this.onTargetCountryChange = this.onTargetCountryChange.bind(this)
    this.onTargetPlatformChange = this.onTargetPlatformChange.bind(this)
    this.onDurationChange = this.onDurationChange.bind(this)
    this.onRepeatChange = this.onRepeatChange.bind(this)
    this.onURLsChange = this.onURLsChange.bind(this)
    this.onCommentChange = this.onCommentChange.bind(this)
    this.onSubmit = this.onSubmit.bind(this)
    this.onEdit = this.onEdit.bind(this)
    this.onAdd = this.onAdd.bind(this)
  }

  static async getInitialProps ({req, res}) {
    // XXX get these from an API call
    const cat_codes = require('../../../static/category-codes.json')
    const countries_alpha2 = require('../../../static/countries-alpha2.json')

    let props = {}
    props.categories = []
    for (let code in cat_codes) {
      props.categories.push({ 'value': code, 'label': cat_codes[code] })
    }

    props.tests = [
      { 'value': 'web_connectivity', 'label': 'Web Connectivity' },
      { 'value': 'http_invalid_request_line', 'label': 'HTTP Invalid Request Line' },
      { 'value': 'http_header_field_manipulation', 'label': 'HTTP Header Field Manipulation' }
    ]

    props.countries = []
    for (let alpha2 in countries_alpha2) {
      props.countries.push({ 'value': alpha2, 'label': countries_alpha2[alpha2] })
    }
    props.countries.sort((a, b) => (+(a.label > b.label) || +(a.label === b.label) - 1))

    props.platforms = [
      { 'value': 'any', 'label': 'any' },
      { 'value': 'android', 'label': 'Android' },
      { 'value': 'ios', 'label': 'iOS' },
      { 'value': 'macos', 'label': 'macOS' },
      { 'value': 'linux', 'label': 'Linux' },
      { 'value': 'lepidopter', 'label': 'Lepidopter' }
    ]
    return props
  }

  componentDidMount() {
    this.state.session.redirectIfInvalid()
  }

  onCommentChange (value) {
    this.setState({ comment: value })
  }

  onURLsChange (value) {
    this.setState({ urls: value });
  }

  onDurationChange (value) {
    this.setState({ duration: value });
  }

  onRepeatChange (value) {
    this.setState({ repeatCount: value });
  }

  onCountryCategoryChange (value) {
    this.setState({ countryCategories: value })
  }

  onGlobalCategoryChange (value) {
    this.setState({ globalCategories: value })
  }

  onTargetCountryChange (value) {
    this.setState({ targetCountries: value })
  }

  onTargetPlatformChange (value) {
    this.setState({ targetPlatforms: value })
  }

  onTestChange (value) {
    if (value.value === 'web_connectivity') {
      this.setState({ inputSelectorOpen: true })
    } else {
      this.setState({ inputSelectorOpen: false })
    }
    this.setState({ selectedTest: value })
  }

  onSubmit () {
    this.setState({
      submitted: true
    })
  }

  onEdit () {
    this.setState({
      submitted: false
    })
  }

  onAdd () {
    let req = this.state.session.createRequest({baseURL: process.env.EVENTS_URL})
    let task_arguments = {}
    let platforms = this.state.targetPlatforms.map((platform) => (platform.value))
    if (platforms.indexOf('any') !== -1) {
      platforms = []
    }
    if (this.state.selectedTest == 'web_connectivity') {
      task_arguments['global_categories'] = this.state.globalCategories.map((category) => (category.value))
      task_arguments['country_categories'] = this.state.countryCategories.map((category) => (category.value))
    }
    req.post('/api/v1/admin/job', {
      'schedule': ToScheduleString({
                      duration: this.state.duration,
                      startMoment: this.state.startMoment,
                      repeatCount: this.state.repeatCount
                  }),
      // XXX we currently don't set this
      'delay': 0,
      'comment': this.state.comment,
      'task': {
        'test_name': this.state.selectedTest.value,
        'arguments': task_arguments,
      },
      'target': {
        'countries': this.state.targetCountries.map((country) => (country.value)),
        'platforms': platforms
      }
    }).then((res) => {
      this.setState({
        finalized: {
          error: null
        },
        submitted: true
      })
      Router.push('/admin/jobs')
    }).catch((err) => {
      this.setState({
        finalized: {
          error: err
        },
        submitted: true
      })
    })
  }

  render () {
    const {
      submitted,
      startDate, startTime,
      startMoment,
      repeatCount,
      globalCategories,
      countryCategories,
      selectedTest,
      targetCountries,
      targetPlatforms,
      duration,
      urls,
      comment,
      finalized
    } = this.state


    return (
      <Layout title="Add Jobs">
        <Head>
          <title>Jobs - OONI Proteus</title>
          <link href="/static/vendor/react-select.css" rel="stylesheet" />
        </Head>

        <div>
          <div className='container'>
            {submitted &&
              <Card>
              <JobCreateConfirm
                startMoment={this.state.startMoment}
                duration={this.state.duration}
                repeatCount={this.state.repeatCount}
                duration={this.state.duration}
                globalCategories={this.state.globalCategories}
                countryCategories={this.state.countryCategories}
                selectedTest={this.state.selectedTest}
                targetCountries={this.state.targetCountries}
                targetPlatforms={this.state.targetPlatforms}
                urls={this.state.urls}
                comment={this.state.comment}
              />
              {!finalized &&
                <CardActions>
                <RaisedButton
                  onTouchTap={this.onEdit}
                  label='Edit'/>
                <RaisedButton
                  onTouchTap={this.onAdd}
                  label='Add'/>
                </CardActions>
              }
              {finalized && finalized.error === null &&
                <p>Job created!</p>}
              {finalized && finalized.error !== null &&
                <div>
                <p>Job creation failed: {finalized.error.toString()}</p>
                <CardActions>
                <RaisedButton
                  onTouchTap={this.onEdit}
                  label='Edit'/>
                <RaisedButton
                  onTouchTap={this.onAdd}
                  label='Retry'/>
                </CardActions>
                </div>
              }
              </Card>
            }
          </div>
          {!submitted &&
          <div className='scheduled-jobs container'>
            <div>

            <div className='section'>
              <h2>Experiment</h2>

              <Grid col={2} px={2}>
              <div className='option'>
                <span className='option-name'>
                  Test
                </span>
                <Select
                  name='test'
                  options={this.props.tests}
                  value={this.state.selectedTest}
                  onChange={this.onTestChange}
                />
              </div>
              </Grid>

              {this.state.inputSelectorOpen
              && <div className='input-selector'>

                <Grid col={3} px={2}>
                  <TextField
                    hintText={`http://example.com/one\nhttp://example.com/two`}
                    floatingLabelText='URLs'
                    multiLine={true}
                    name="urls"
                    value={this.state.urls}
                    onChange={(event, value) => this.onURLsChange(value)}
                    rows={3}
                  />
                </Grid>

                <Grid col={3} px={2}>
                  <div className='option'>
                    <span className='option-name'>
                      Global Categories
                    </span>
                    <Select
                      multi
                      name='global-categories'
                      options={this.props.categories}
                      value={this.state.globalCategories}
                      onChange={this.onGlobalCategoryChange}
                    />
                  </div>
                </Grid>

                <Grid col={3} px={2}>
                  <div className='option'>
                    <span className='option-name'>
                      Country Categories
                    </span>

                    <Select
                      multi
                      name='country-categories'
                      value={this.state.countryCategories}
                      onChange={this.onCountryCategoryChange}
                      options={this.props.categories}
                    />
                  </div>
                </Grid>
              </div>}

            </div>

            <div className='section'>
              <h2>Schedule</h2>
              <Flex>
              <Box px={2}>
                <div className='option'>
                  <span className='option-name'>
                    Start on
                  </span>

                  <DatePicker
                    floatingLabelText="Start date"
                    autoOk={true}
                    value={this.state.startDate}
                    onChange={(event, startDate) => {
                      let startMoment = this.state.startMoment.clone()
                      // XXX is it correct to use UTC here?
                      startMoment.set({
                        year: startDate.getUTCFullYear(),
                        month: startDate.getUTCMonth(),
                        date: startDate.getUTCDate(),
                      })
                      this.setState({ startMoment })
                      this.setState({ startDate })
                    }}
                  />
                  <TimePicker
                    format="24hr"
                    value={this.state.startTime}
                    hintText="Start time"
                    onChange={(event, startTime) => {
                      let startMoment = this.state.startMoment.clone()
                      // XXX is it correct to use UTC here?
                      startMoment.set({
                        hour: startTime.getUTCHours(),
                        minute: startTime.getUTCMinutes(),
                        second: startTime.getUTCSeconds(),
                      })
                      this.setState({ startMoment })
                      this.setState({ startTime })
                    }}
                  />
                  <RaisedButton
                    onTouchTap={() => {this.setState({startMoment: moment(new Date()), startTime: new Date(), startDate: new Date()})}}
                    label='Now'/>
                </div>
              </Box>

              <Box px={2}>
              <div className='option'>
                <span className='option-name'>
                  Repeat
                </span>
                <RepeatString duration={this.state.duration} repeatCount={this.state.repeatCount} />
                <DurationPicker onChange={this.onDurationChange} />
                <Checkbox
                  label="Repeat forever"
                  checked={this.state.repeatCount === 0}
                  onCheck={(event, isInputChecked) => {
                    if (isInputChecked === true) this.onRepeatChange(0)
                    else this.onRepeatChange(1)
                  }}
                />
                <TextField
                  style={{width: 20, float: 'left'}}
                  name='repeat-count'
                  value={this.state.repeatCount}
                  onChange={(event, value) => {this.onRepeatChange(value)}}
                />
                <Slider
                  style={{width: 100, float: 'left', marginLeft: 20}}
                  min={1}
                  max={99}
                  step={1}
                  disabled={this.state.repeatCount === 0}
                  value={this.state.repeatCount}
                  defaultValue={1}
                  onChange={(event, value) => {this.onRepeatChange(value)}}
                />
              </div>
              </Box>
              </Flex>

            </div>

            <div className='section'>
              <h2>Target</h2>
              <Grid col={3} px={2}>
              <div className='option'>
                <span className='option-name'>
                  Country
                </span>
                <Select
                  multi
                  name='target-country'
                  options={this.props.countries}
                  value={this.state.targetCountries}
                  onChange={this.onTargetCountryChange}
                />
              </div>
              </Grid>

              <Grid col={3} px={2}>
              <div className='option'>
                <span className='option-name'>
                  Platform
                </span>
                <Select
                  multi
                  name='target-platform'
                  options={this.props.platforms}
                  value={this.state.targetPlatforms}
                  onChange={this.onTargetPlatformChange}
                />
              </div>
              </Grid>

            </div>

            <div className='section'>
              <h2>Submit</h2>
              <Grid col={6} px={2}>
              <TextField
                hintText='make it something descriptive'
                name='task-comment'
                floatingLabelText='Task comment'
                value={this.state.comment}
                onChange={(event, value) => this.onCommentChange(value)}
              />
              <RaisedButton
                onTouchTap={this.onSubmit}
                label='Add' style={{marginLeft: 20}}/>
              </Grid>

            </div>

          </div>
          </div>}
          <style jsx>{`
          .container {
            max-width: 1024px;
            padding-left: 20px;
            padding-right: 20px;
            margin: auto;
          }
          .section {
            padding: 20px;
          }
          .section h2 {
            padding-bottom: 20px;
            font-weight: 100;
          }
          .option-name {
            display: block;
            padding-bottom: 10px;
          }
          .option {
            padding-bottom: 20px;
          }
          .url-list {
            width: 300px;
            min-height: 100px;
          }
          `}</style>
        </div>
      </Layout>
    )
  }
}
