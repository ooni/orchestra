export default class Session {

  constructor({req} = {}) {
    this._session = {}
    try {
      if (req) {
        // Server-side logic
        this._session = {}
      } else {
        this._session = this._getLocalStore('session')
      }
    } catch (err) {
      // fallthrough
      return
    }
  }

  isValid() {
    if (this._session && Object.keys(this._session).length > 0 && this._session.expires && this._session.expires - Date.now() >= 0) {
      return true
    }
    return false
  }

  async getSession() {
    if (typeof window === 'undefined') {
      return new Promise(resolve => {
        resolve(this._session)
      })
    }

    this._session = this._getLocalStore('session')
    return new Promise(resolve => {
      resolve(this._session)
    })
  }

  async login(username, password) {
    return new Promise(async (resolve, reject) => {
      if (typeof window === 'undefined') {
        return reject(Error('This method is called only in the client'))
      }
      let xhr = new XMLHttpRequest()
      xhr.open('POST', process.env.REGISTRY_URL + '/api/v1/login')
      xhr.setRequestHeader('Content-type', 'application/json')
      xhr.onreadystatechange = async () => {
        if (xhr.readyState === 4) {
          if (xhr.status !== 200) {
            return reject(Error('XMLHttpRequest error: error while logging in'))
          }
          this._session = JSON.parse(xhr.responseText)
          this._session.expire = new Date(this._session.expire)
          this._session.username = username
          this._saveLocalStore('session', this._session)

          return resolve(true)
        }
      }
      xhr.onerror = () => {
        return reject(Error('XMLHttpRequest error: unable to login'))
      }
      xhr.send(JSON.stringify({username, password}))
    })
  }

  _getLocalStore(name) {
    try {
      return JSON.parse(localStorage.getItem(name))
    } catch (err) {
      return null
    }
  }

  _saveLocalStore(name, data) {
    try {
      localStorage.setItem(name, JSON.stringify(data))
      return true
    } catch (err) {
      return false
    }
  }

  _removeLocalStore(name) {
    try {
      localStorage.removeItem(name)
      return true
    } catch(err) {
      return false
    }
  }
}
