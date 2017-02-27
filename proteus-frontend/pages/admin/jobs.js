import React from 'react'
import Head from 'next/head'

import NoSSR from 'react-no-ssr'
import Select from 'react-select'

import axios from 'axios'
import Immutable from 'immutable'
import moment from 'moment'

import RaisedButton from 'material-ui/RaisedButton'
import Checkbox from 'material-ui/Checkbox'
import DatePicker from 'material-ui/DatePicker'
import TimePicker from 'material-ui/TimePicker'
import Slider from 'material-ui/Slider'
import TextField from 'material-ui/TextField'

import Layout from '../../components/layout'

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
          return <span>{value} {unitName} </span>
        }
      })}
      <style jsx>{`
      div {
        padding-bottom: 20px;
      `}</style>
    </div>
  )
}

// XXX protect this with some auth
export default class AdminJobs extends React.Component {

  constructor(props) {
    super(props)
    this.state = {
      startDate: null,
      startTime: null,
      repeatCount: 0,
      globalCategories: [],
      countryCategories: [],
      selectedTest: {},
      targetCountries: [],
      targetPlatforms: [],
      duration: {},
      urls: '',
      inputSelectorOpen: false,
      comment: ''
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
  }

  static async getInitialProps () {
		// XXX get this from lists API server
    let categories = [{'value': 'ALDR', 'label': 'Alcohol & Drugs'}, {'value': 'REL', 'label': 'Religion'}, {'value': 'PORN', 'label': 'Pornography'}, {'value': 'PROV', 'label': 'Provocative Attire'}, {'value': 'POLR', 'label': 'Political Criticism'}, {'value': 'HUMR', 'label': 'Human Rights Issues'}, {'value': 'ENV', 'label': 'Environment'}, {'value': 'MILX', 'label': 'Terrorism and Militants'}, {'value': 'HATE', 'label': 'Hate Speech'}, {'value': 'NEWS', 'label': 'News Media'}, {'value': 'XED', 'label': 'Sex Education'}, {'value': 'PUBH', 'label': 'Public Health'}, {'value': 'GMB', 'label': 'Gambling'}, {'value': 'ANON', 'label': 'Anonymization and circumvention tools'}, {'value': 'DATE', 'label': 'Online Dating'}, {'value': 'GRP', 'label': 'Social Networking'}, {'value': 'LGBT', 'label': 'LGBT'}, {'value': 'FILE', 'label': 'File-sharing'}, {'value': 'HACK', 'label': 'Hacking Tools'}, {'value': 'COMT', 'label': 'Communication Tools'}, {'value': 'MMED', 'label': 'Media sharing'}, {'value': 'HOST', 'label': 'Hosting and Blogging Platforms'}, {'value': 'SRCH', 'label': 'Search Engines'}, {'value': 'GAME', 'label': 'Gaming'}, {'value': 'CULTR', 'label': 'Culture'}, {'value': 'ECON', 'label': 'Economics'}, {'value': 'GOVT', 'label': 'Government'}, {'value': 'COMM', 'label': 'E-commerce'}, {'value': 'CTRL', 'label': 'Control content'}, {'value': 'IGO', 'label': 'Intergovernmental Organizations'}, {'value': 'MISC', 'label': 'Miscelaneous content'}]

    let tests = [
      { 'value': 'web_connectivity', 'label': 'Web Connectivity' },
      { 'value': 'http_invalid_request_line', 'label': 'HTTP Invalid Request Line' },
      { 'value': 'http_header_field_manipulation', 'label': 'HTTP Header Field Manipulation' }
    ]

    const countries_alpha2 = require('../../static/countries-alpha2.json')
    let countries = []
    for (let alpha2 in countries_alpha2) {
      countries.push({ 'value': alpha2, 'label': countries_alpha2[alpha2] })
    }
    countries.sort((a, b) => (+(a.label > b.label) || +(a.label === b.label) - 1))

    let platforms = [
      { 'value': 'any', 'label': 'any' },
      { 'value': 'android', 'label': 'Android' },
      { 'value': 'ios', 'label': 'iOS' },
      { 'value': 'macos', 'label': 'macOS' },
      { 'value': 'linux', 'label': 'Linux' },
      { 'value': 'lepidopter', 'label': 'Lepidopter' }
    ]

    return { categories, tests, countries, platforms }
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

  render () {
    return (
      <Layout>
        <Head>
          <title>Jobs - OONI Proteus</title>
          <link href="/static/vendor/_datepicker.css" rel="stylesheet" />
          <link href="/static/vendor/timepicker.css" rel="stylesheet" />
          <link href="/static/vendor/react-select.css" rel="stylesheet" />
        </Head>
        <div className='scheduled-jobs'>
          <h1>Add periodic job</h1>
          <div>

          <div className='section'>
            <h2>Schedule</h2>
            <div className='option'>
              <span className='option-name'>
                Start on
              </span>

							<DatePicker
								floatingLabelText="Start date"
								autoOk={true}
								value={this.state.startDate}
								onChange={(event, startDate) => { this.setState({ startDate })} }
							/>
              <TimePicker
                format="24hr"
                value={this.state.startTime}
                hintText="Start time"
								onChange={(event, startTime) => { this.setState({ startTime })} }
              />

            </div>

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

          </div>

          <div className='section'>
            <h2>Experiment</h2>

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

            {this.state.inputSelectorOpen
            && <div className='input-selector'>
                <TextField
                  hintText={`http://example.com/one\nhttp://example.com/two`}
                  floatingLabelText='URLs'
                  multiLine={true}
                  value={this.state.urls}
                  onChange={(event, value) => this.onURLsChange(value)}
                  rows={3}
                />

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
            </div>}

          </div>

          <div className='section'>
            <h2>Target</h2>

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

          </div>

          <div className='section'>
            <h2>Submit</h2>
            <TextField
              hintText='make it something descriptive'
              floatingLabelText='Task comment'
              value={this.state.comment}
              onChange={(event, value) => this.onCommentChange(value)}
            />

            <RaisedButton label='Add' style={{marginLeft: 20}}/>
          </div>

          </div>
          <style jsx>{`
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
