import Router from 'next/router'

import Session from '../../components/session'
import Layout from '../../components/layout'

import { withStyles } from '@material-ui/styles'

import Button from '@material-ui/core/Button'
import Input from '@material-ui/core/Input'

import Card, { CardHeader, CardContent, CardActions } from '@material-ui/core/Card'

import { Flex, Box, Grid } from 'ooni-components'

class AdminLogin extends React.Component {

  constructor(props) {
    super(props)
    this.state = {
      username: '',
      password: '',
      error: null
    }
    this.onSubmit = this.onSubmit.bind(this)
    this.onUsernameChange = this.onUsernameChange.bind(this)
    this.onPasswordChange = this.onPasswordChange.bind(this)
  }

  onUsernameChange(username) {
    this.setState({username})
  }
  onPasswordChange(password) {
    this.setState({password})
  }

  async onSubmit(event) {
    event.preventDefault()
    const session = new Session()
    session.login(this.state.username, this.state.password)
      .then(() => {
        let redirectPath = '/'
        this.setState({
          error: null
        })
        if (Router.query.from) {
          redirectPath = Router.query['from']
        }
        Router.push(redirectPath)
      })
      .catch(err => {
        this.setState({
          error: {
            'message': 'Failed to login',
            'debug_error': err
          }
        })
      })
  }

  render() {
    const {
      error
    } = this.state

    const { classes } = this.props
    return (
      <Layout>
        <Card style={{width: "600px", margin: "auto", padding: '20px'}}>
          {error !== null && <p>{error.message}</p>}
          <CardHeader title="Authorization required" style={{color: "black"}} />
          <CardContent>
          <Input
            className={classes.input}
            placeholder='Username'
            value={this.state.username}
            onChange={({target}) => this.onUsernameChange(target.value)}
          />
          <Input
            className={classes.input}
            type='password'
            placeholder='Password'
            value={this.state.password}
            onChange={({target}) => this.onPasswordChange(target.value)}
            />
          </CardContent>
          <CardActions>
          <Button className={classes.button} onClick={this.onSubmit}>Login</Button>
          </CardActions>
        </Card>
      </Layout>
    )
  }
}

const styles = theme => ({
  container: {
    display: 'flex',
    flexWrap: 'wrap',
  },
  button: {},
  input: {
    margin: theme.spacing.unit,
  },
})

export default withStyles(styles)(AdminLogin)
