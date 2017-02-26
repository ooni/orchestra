import React from 'react'
import Head from 'next/head'

import NoSSR from 'react-no-ssr'
import Select from 'react-select'

import axios from 'axios'
import Immutable from 'immutable'
import moment from 'moment'

import DatePicker from 'material-ui/DatePicker'
import TimePicker from 'material-ui/TimePicker'

import Layout from '../../components/layout'

class RepeatPicker extends React.Component {
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
    this.setState({duration: this.state.duration.set(unit, quantity)})
  }

  toggleOpen () {
    this.setState({ isOpen: !this.state.isOpen })
  }

  render () {
    const self = this
    const makeItem = (unit) => {
      if (this.state.duration.get(unit) === 0) {
        return null
      }
      return (
        <div className='item'>
          <span className='quantity'>
            {this.state.duration.get(unit)}
          </span>
          <span className='unit'>
            {unit}
          </span>
          <style jsx>{`
          .item {
            display: inline;
            padding-right: 2px;
          }
          .item .quantity {
            font-size: 16px;
            font-weight: 200;
          }
          .item .unit {
            font-size: 14px;
            font-weight: 100;
          }
          `}</style>
        </div>
      )
    }

    const makeSelector = (unit, range) => {
      const options = Array.apply(0, Array(range))
        .map((_, idx) => {
          return <div className='option'
                      key={idx}
                      onClick={() => { self.setDuration(unit, idx) } }>
            {idx}
            <style jsx>{`
            .option {
              padding: 0 10px;
            }
            .option:hover {
              background-color: #ccc;
            }
            `}</style>
          </div>
        })

      return (
        <div className='selector-options'>
          <div className='unit'>
            {unit}
          </div>
          <div className='options'>
            {options}
          </div>
          <style jsx>{`
          .unit {
            padding: 10px;
          }
          .options {
            height: 100px;
            padding-right: 10px;
            overflow-y: scroll;
          }
          .selector-options {
            float: left;
          }
          `}</style>
        </div>
      )
    }

    return (
      <div className='picker'>
        <div className='current-selection' onClick={() => this.toggleOpen() }>
          {makeItem('Y')}
          {makeItem('M')}
          {makeItem('W')}
          {makeItem('D')}
          {makeItem('h')}
          {makeItem('m')}
          {makeItem('s')}
        </div>
        {this.state.isOpen
        && <div className='selector'>
          {makeSelector('Y', 10)}
          {makeSelector('M', 12)}
          {makeSelector('W', 5)}
          {makeSelector('D', 30)}
          {makeSelector('h', 24)}
          {makeSelector('m', 60)}
          {makeSelector('s', 60)}
        </div>}
        <style jsx>{`
          .picker {
            position: relative;
          }
          .current-selection {
            padding: 10px;
            color: #000;
            background-color: #fff;
          }
          .selector {
            position: absolute;
            left: 0;
            color: #000;
            background-color: #fff;
          }
        `}</style>
      </div>
    )
  }

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
      inputSelectorOpen: false
    }

    this.onRepeatChange = this.onRepeatChange.bind(this);
    this.onCountryCategoryChange = this.onCountryCategoryChange.bind(this);
    this.onGlobalCategoryChange = this.onGlobalCategoryChange.bind(this);
    this.onTestChange = this.onTestChange.bind(this);
    this.onTargetCountryChange = this.onTargetCountryChange.bind(this);
    this.onTargetPlatformChange = this.onTargetPlatformChange.bind(this);
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

  onRepeatChange (event) {
    this.setState({ repeatCount: event.target.value });
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
              <RepeatPicker />
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
              <div className='option'>
                <span className='option-name'>
                  URLS
                </span>
                <textarea className='url-list'></textarea>
              </div>

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

            <div className='option'>
              <span className='option-name'>
                Task comment
              </span>
              <input type='text' />
            </div>

            <input type='submit' value='Add' />
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
