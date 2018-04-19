import React from 'react'
import PropTypes from 'prop-types'
import Head from 'next/head'
import Router from 'next/router'

import Select from 'react-select'
//import Autocomplete from 'react-toolbox/lib/autocomplete/Autocomplete'

import Button from 'material-ui/Button'
//import Button from 'react-toolbox/lib/button/Button'
//import Input from 'react-toolbox/lib/input/Input'
import Input from 'material-ui/Input'

//import Checkbox from 'react-toolbox/lib/checkbox/Checkbox'
import Checkbox from 'material-ui/Checkbox'
import { FormControlLabel } from 'material-ui/Form'

import Card, { CardHeader, CardContent, CardActions } from 'material-ui/Card'
//import Card from 'react-toolbox/lib/card/Card'
//import CardTitle from 'react-toolbox/lib/card/CardTitle'
//import CardActions from 'react-toolbox/lib/card/CardActions'
//import CardText from 'react-toolbox/lib/card/CardText'

//import DatePicker from 'react-toolbox/lib/date_picker/DatePicker'
//import TimePicker from 'react-toolbox/lib/time_picker/TimePicker'

import TextField from 'material-ui/TextField'

import List, { ListItem } from 'material-ui/List'
//import List from 'react-toolbox/lib/list/List'
//import ListItem from 'react-toolbox/lib/list/ListItem'

import moment from 'moment'

import Layout from '../../../components/layout'
import Session from '../../../components/session'
import {
  DesignatorSlider,
  DurationPicker,
  RepeatString,
  ToScheduleString
} from '../../../components/ui/schedule'

import { Flex, Box, Grid } from 'ooni-components'
import { Container } from 'ooni-components'


class JobCreateConfirm extends React.Component {
  static propTypes = {
    startMoment: PropTypes.object,
    duration: PropTypes.object,
    repeatCount: PropTypes.number,
    targetCountries: PropTypes.array,
    targetPlatforms: PropTypes.array,
    urls: PropTypes.string,
    alertMessage: PropTypes.string,
    href: PropTypes.string,
    altHref: PropTypes.string
  }

  constructor(props) {
    super(props)
  }

  render() {
    const {
      startMoment,
      duration,
      repeatCount,
      alertMessage,
      href,
      altHref,
      targetCountries,
      targetPlatforms
    } = this.props

    const DurationCaption = ({duration, repeatCount, startMoment}) => (<div>
      <RepeatString duration={duration} repeatCount={repeatCount} />
      {ToScheduleString({
        duration: duration,
        startMoment: startMoment,
        repeatCount: repeatCount
      })}
    </div>
    )

    let startTimeCaption = startMoment.calendar()
    startTimeCaption += ' ('
    startTimeCaption += startMoment.toString()
    startTimeCaption += ')'

    return (
      <div>
        <CardHeader title="New Alert Summary" />
        <CardContent>
          <List>
          <ListItem
            ripple={false}
            primary={alertMessage}
            secondary="Message" />
          <ListItem
            ripple={false}
            primary={href}
            secondary="Link" />
          <ListItem
            ripple={false}
            primary={startTimeCaption}
            secondary="Start time" />
          <ListItem
            ripple={false}
            primary={<DurationCaption
                        duration={duration}
                        repeatCount={repeatCount}
                        startMoment={startMoment} />}
            secondary="Duration" />

          <ListItem
            ripple={false}
            primary={targetCountries.join(',')}
            secondary="Target pountries" />
          <ListItem
            ripple={false}
            primary={targetPlatforms.join(',')}
            secondary="Target platforms" />
          </List>

        </CardContent>

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
      repeatCount: 1,
      alertMessage: '',
      href: '',
      altHref: '',
      targetCountries: [],
      targetPlatforms: [],
      duration: {W: 1},
      inputSelectorOpen: false,
      submitted: false,
      comment: '',
      session: new Session(),
      finalized: null
    }

    this.onTargetCountryChange = this.onTargetCountryChange.bind(this)
    this.onTargetPlatformChange = this.onTargetPlatformChange.bind(this)
    this.onDurationChange = this.onDurationChange.bind(this)
    this.onRepeatChange = this.onRepeatChange.bind(this)
    this.onSubmit = this.onSubmit.bind(this)
    this.onEdit = this.onEdit.bind(this)
    this.onAdd = this.onAdd.bind(this)
    this.onMessageChange = this.onMessageChange.bind(this)
    this.onHrefChange = this.onHrefChange.bind(this)
    this.onAltHrefChange = this.onAltHrefChange.bind(this)
  }

  static async getInitialProps ({req, res}) {
    // XXX get these from an API call
    const cat_codes = require('../../../static/category-codes.json')
    const countries_alpha2 = require('../../../static/countries-alpha2.json')

    let props = {}
    props.countries = [
      {value: "any", label: 'All'}
    ]
    for (let alpha2 in countries_alpha2) {
      props.countries.push({
        value: alpha2,
        label: countries_alpha2[alpha2]
      })
    }

    props.platforms = [
      {value: 'any', label: 'All'},
      {value: 'android', label: 'Android'},
      {value: 'ios', label: 'iOS'},
      {value: 'linux', label: 'Linux'},
      {value: 'macos', label: 'macOS'},
      {value: 'lepidopter', label: 'Lepidopter'},
    ]

    return props
  }

  componentDidMount() {
    this.state.session.redirectIfInvalid()
  }

  onMessageChange ({target}) {
    this.setState({ alertMessage: target.value})
  }
  onHrefChange ({target}) {
    this.setState({ href: target.value})
  }
  onAltHrefChange ({target}) {
    this.setState({ altHref: target.value})
  }

  onDurationChange ({target}) {
    this.setState({ duration: target.value });
  }

  onRepeatChange ({target}) {
    this.setState({ repeatCount: target.value });
  }

  onTargetCountryChange (value) {
    if (value.indexOf('any') != -1) {
      if (this.state.targetCountries.indexOf('any') != -1) {
        // If any was already there we remove it
        value.pop('any')
      } else {
        // Otherwise we clear everything else
        value = ['any']
      }
    }
    this.setState({ targetCountries: value })
  }

  onTargetPlatformChange (value) {
    if (value.indexOf('any') != -1) {
      if (this.state.targetPlatforms.indexOf('any') != -1) {
        // If any was already there we remove it
        value.pop('any')
      } else {
        // Otherwise we clear everything else
        value = ['any']
      }
    }
    this.setState({ targetPlatforms: value })
  }

  onSubmit () {
    console.log('on submit')
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
    console.log('on add')
    let req = this.state.session.createRequest({baseURL: process.env.ORCHESTRATE_URL})
    let alertExtra = {}
    if (this.state.href != '') {
      alertExtra['href'] = this.state.href
      alertExtra['alt_hrefs'] = []
      if (this.state.altHref != '') {
        this.state.altHref.split("\n").forEach((v) => {
          if (v != '') {
            alertExtra['alt_hrefs'].push(v)
          }
        })
      }
    }
    let platforms = this.state.targetPlatforms.slice()
    if (platforms.indexOf('any') !== -1) {
      platforms = []
    }
    let countries = this.state.targetCountries.slice()
    if (countries.indexOf('any') !== -1) {
      countries = []
    }
    req.post('/api/v1/admin/job', {
      'schedule': ToScheduleString({
                      duration: this.state.duration,
                      startMoment: this.state.startMoment,
                      repeatCount: this.state.repeatCount
                  }),
      // XXX we currently don't set this
      'delay': 0,
      'comment': this.state.alertMessage,
      'alert': {
        'message': this.state.alertMessage,
        'extra': alertExtra,
      },
      'target': {
        'countries': countries,
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
      alertMessage,
      targetCountries,
      targetPlatforms,
      duration,
      finalized
    } = this.state


    return (
      <Layout title="Add Jobs">
        <Head>
          <title>Jobs - OONI Proteus</title>
          <link href="/static/vendor/react-select.css" rel="stylesheet" />
        </Head>

        <div>
          <Container>
            {submitted &&
              <Card>
              {finalized && finalized.error === null &&
                <CardHeader title="Job created successfully!" />
              }
              {finalized && finalized.error !== null &&
                <div>
                <CardHeader title="Job creation error" />
                <p>{finalized.error.toString()}</p>
                <CardActions>
                <Button
                  raised
                  onClick={this.onEdit}>Edit</Button>
                <Button
                  raised
                  onClick={this.onAdd}>Retry</Button>
                </CardActions>
                </div>
              }
              {!finalized && <div>
                <JobCreateConfirm
                  startMoment={this.state.startMoment}
                  duration={this.state.duration}
                  repeatCount={this.state.repeatCount}
                  duration={this.state.duration}

                  alertMessage={this.state.alertMessage}
                  href={this.state.href}
                  altHref={this.state.altHref}

                  targetCountries={this.state.targetCountries}
                  targetPlatforms={this.state.targetPlatforms}
                  urls={this.state.urls}
                  comment={this.state.comment}
                />
                <CardActions>
                <Button
                  onClick={this.onEdit}>Edit</Button>
                <Button
                  onClick={this.onAdd}>Add</Button>
                </CardActions>
              </div>}

              </Card>
            }
          </Container>
          {!submitted &&
          <Container>
            <Card title="New Alert">
              <CardContent>
                <Input
                  onChange={this.onMessageChange}
                  placeholder="message"
                  type="text" />
                <Input
                  onChange={this.onHrefChange}
                  placeholder="href"
                  type="text" />
                <Input
                  onChange={this.onAltHrefChange}
                  placeholder="alt hrefs"
                  multiline
                  type="text" />
              <hr/>

              <h2>Target</h2>

              <div className='option'>
                <span className='option-name'>
                  Country
                </span>
                <Select
                  name='test'
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
                  name='test'
                  options={this.props.platforms}
                  value={this.state.targetPlatforms}
                  onChange={this.onTargetPlatformChange}
                />
              </div>

              <hr />
              <h2>Schedule</h2>
              <Flex>
              <Box px={2}>
                <div className='option'>
                  <span className='option-name'>
                    Start on
                  </span>

                  <TextField
                    id="date"
                    label="Start time"
                    type="datetime-local"
                    value={this.state.startTime}
                    onChange={({target}) => {
                      let startMoment = moment(target.value)
                      this.setState({ startMoment })
                      this.setState({ startDate: startMoment.toDate() })
                      this.setState({ startTime: startMoment.toDate() })
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
                    >Now</Button>
                </div>
              </Box>

              <Box px={2}>
                <FormControlLabel
                  control={
                    <Checkbox
                    checked={this.state.repeatCount !== 1}
                    onChange={(isInputChecked) => {
                      console.log(isInputChecked)
                      if (isInputChecked === true) this.onRepeatChange(2)
                      else this.onRepeatChange(1)
                    }}
                    />
                  }
                  label="Repeat"
                  />
                {this.state.repeatCount !== 1 && <div>
                  <FormControlLabel
                  control={
                    <Checkbox
                    checked={this.state.repeatCount === 0}
                    onChange={(isInputChecked) => {
                      if (isInputChecked === true) this.onRepeatChange(0)
                      else this.onRepeatChange(2)
                    }}
                  />
                  }
                  label="Repeat forever"
                  />

                  {this.state.repeatCount !== 0 && <Input
                  type='number'
                  min='0'
                  placeholder='times to repeat'
                  name='repeat-count'
                  value={this.state.repeatCount}
                  onChange={(value) => {this.onRepeatChange(value)}}
                  />}
                <DurationPicker onChange={this.onDurationChange}
                                duration={this.state.duration} />

                <RepeatString duration={this.state.duration} repeatCount={this.state.repeatCount} />
                </div>}
              </Box>
              </Flex>

              </CardContent>
              <CardActions>
                <Button
                  raised
                  onClick={this.onSubmit}
                  style={{marginLeft: 20}}>Add</Button>
              </CardActions>
            </Card>

          </Container>}
        </div>
      </Layout>
    )
  }
}
