import React from 'react'
import Head from 'next/head'
import Router from 'next/router'

import Select from 'react-select'

import Immutable from 'immutable'

import Button from 'react-toolbox/lib/button/Button'
import Input from 'react-toolbox/lib/input/Input'
import Checkbox from 'react-toolbox/lib/checkbox/Checkbox'
import Slider from 'react-toolbox/lib/slider/Slider'

import Card from 'react-toolbox/lib/card/Card'
import CardTitle from 'react-toolbox/lib/card/CardTitle'
import CardActions from 'react-toolbox/lib/card/CardActions'
import CardText from 'react-toolbox/lib/card/CardText'

import DatePicker from 'react-toolbox/lib/date_picker/DatePicker'
import TimePicker from 'react-toolbox/lib/time_picker/TimePicker'

import moment from 'moment'

import Layout from '../../../components/layout'
import Session from '../../../components/session'
import {
  DesignatorSlider,
  DurationPicker,
  RepeatString,
  ToScheduleString
} from '../../../components/ui/schedule'

import { Flex, Box, Grid } from 'reflexbox'

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
      startDate: new Date(),
      startTime: new Date(),
      startMoment: moment(),
      repeatCount: 0,
      globalCategories: [],
      countryCategories: [],
      selectedTest: {},
      targetCountries: [],
      targetPlatforms: [],
      duration: {m: 10},
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
    let req = this.state.session.createRequest({baseURL: process.env.ORCHESTRATE_URL})
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
      startDate,
      startTime,
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
                <Button
                  onClick={this.onEdit}
                  label='Edit'/>
                <Button
                  onClick={this.onAdd}
                  label='Add'/>
                </CardActions>
              }
              {finalized && finalized.error === null &&
                <p>Job created!</p>}
              {finalized && finalized.error !== null &&
                <div>
                <p>Job creation failed: {finalized.error.toString()}</p>
                <CardActions>
                <Button
                  raised
                  onClick={this.onEdit}
                  label='Edit'/>
                <Button
                  raised
                  onClick={this.onAdd}
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
                  <Input
                    hint={`http://example.com/one\nhttp://example.com/two`}
                    label='URLs'
                    multiline
                    name="urls"
                    value={this.state.urls}
                    onChange={this.onURLsChange}
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
                    label="Start date"
                    autoOk={true}
                    value={this.state.startDate}
                    onChange={(startDate) => {
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
                    label="Start time"
                    onChange={(startTime) => {
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
                  <Button
                    onClick={() => {
                        this.setState({
                          startMoment: moment(new Date()),
                          startTime: new Date(),
                          startDate: new Date()
                        })
                      }
                    }
                    label='Now'/>
                </div>
              </Box>

              <Box px={2}>
              <div className='option'>
                <span className='option-name'>
                  Repeat
                </span>
                <RepeatString duration={this.state.duration} repeatCount={this.state.repeatCount} />
                <DurationPicker onChange={this.onDurationChange}
                                duration={this.state.duration} />
                <Checkbox
                  label="Repeat forever"
                  checked={this.state.repeatCount === 0}
                  onChange={(isInputChecked) => {
                    if (isInputChecked === true) this.onRepeatChange(0)
                    else this.onRepeatChange(1)
                  }}
                />
                <Input
                  type='text'
                  style={{width: 20, float: 'left'}}
                  name='repeat-count'
                  value={this.state.repeatCount}
                  onChange={(value) => {this.onRepeatChange(value)}}
                />
                <Slider
                  style={{width: 100, float: 'left', marginLeft: 20}}
                  min={1}
                  max={99}
                  step={1}
                  disabled={this.state.repeatCount === 0}
                  value={this.state.repeatCount}
                  onChange={this.onRepeatChange}
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
              <Input
                hintText='make it something descriptive'
                name='task-comment'
                floatingLabelText='Task comment'
                value={this.state.comment}
                onChange={this.onCommentChange}
              />
              <Button
                raised
                onClick={this.onSubmit}
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
