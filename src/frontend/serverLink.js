var userManager;
var drawingManager;

async function postRequest(url, data) {
  let response = await fetch(url,
    {
      method: "POST",
      headers: {"Content-Type": "application/json"},
      body: JSON.stringify(data),
    }
  );
  var result = await response.json();
  result.statusCode = response.status;
  return result;
}

async function patchRequest(url, data) {
  let response = await fetch(url,
    {
      method: "PATCH",
      headers: {"Content-Type": "application/json"},
      body: JSON.stringify(data),
    }
  );
  var result = await response.json();
  result.statusCode = response.status;
  return result;
}

async function getRequest(url) {
  let result = await fetch(url);
  let json = await result.json();
  json.statusCode = result.status;
  return json
}

function getCookie(name) {
  // Cheers https://stackoverflow.com/a/25490531/27835424
  return document.cookie.match('(^|;)\\s*' + name + '\\s*=\\s*([^;]+)')?.pop() || ''
}


class UserManager {
  
  constructor() {
    this.user = null;
  }

  isLoggedin() {
    return this.user != null;
  }

  getUsername() {
    if (this.user.email.length > 0) {
      return this.user.email.substring(0, this.user.email.indexOf("@"));
    }
    return "";
  }

  userOnlyComponents() {
    return [
      bodyComponent.rightMenuComponent.logoutButtonComponent,
      bodyComponent.rightMenuComponent.saveButtonComponent,
      bodyComponent.rightMenuComponent.myDrawingsButtonComponent,
      bodyComponent.rightMenuComponent.welcomeMsgComponent,
    ]
  }
  guestOnlyComponents() {
    return [
      bodyComponent.rightMenuComponent.loginButtonComponent,
      bodyComponent.rightMenuComponent.signupButtonComponent,
    ]
  }

  renderLogin() {
    this.userOnlyComponents().forEach(component => component.show());
    this.guestOnlyComponents().forEach(component => component.hide());
    let username = this.getUsername();
    bodyComponent.rightMenuComponent.welcomeMsgComponent.setValue("Welcome, " + username);
  }
  
  renderLogout() {
    this.userOnlyComponents().forEach(component => component.hide());
    this.guestOnlyComponents().forEach(component => component.show());
  }

  hideAll() {
    // This is used so that components don't briefly appear before API calls
    // on page load.
    this.userOnlyComponents().forEach(component => component.hide());
    this.guestOnlyComponents().forEach(component => component.hide());
  }

  signupCallback(response) {
    if (response.error == "") {
      bodyComponent.informerComponent.report("Successfully signed up!", "good");
      bodyComponent.signupComponent.hide();
      bodyComponent.loginComponent.formComponent.formFieldEmail.setValue(
        bodyComponent.signupComponent.formComponent.formFieldEmail.getValue()
      )
      bodyComponent.loginComponent.formComponent.formFieldPassword.setValue(
        bodyComponent.signupComponent.formComponent.formFieldPassword.getValue()
      )
      bodyComponent.signupComponent.formComponent.formClear();
      return;
    }
    bodyComponent.informerComponent.report(response.error, "bad");
  }

  async loginCallback(response) {
    if (response.error == "") {
      bodyComponent.informerComponent.report("Successfully logged in!", "good");
      bodyComponent.loginComponent.hide();
      bodyComponent.loginComponent.formComponent.formClear();
      await userManager.update();
      return;
    }
    bodyComponent.informerComponent.report(response.error, "bad");
  }

  async update() {
    await this.setUser();
    this.isLoggedin()? this.renderLogin(): this.renderLogout();
  }

  async setUser() {
    let response = await getRequest("/api/user/");
    if (response.id > -1 && response.email.length > 0) {
      this.user = {email: response.email, userId: response.id};
      return
    }
    this.user = null;
  }

  async logout() {
    await getRequest("/api/user/logout");
    await this.update();
    bodyComponent.hidePopups();
    bodyComponent.informerComponent.report("Successfully logged out!", "good");
  }

  async login(data) {
    return await postRequest("/api/user/auth", data);
  }

  async signup(data) {
    return await postRequest("/api/user/", data);
  }
}


class DrawingManager {
  
  constructor() {
    this.currentDrawingId = null; // TODO: Fetch from local storage
  }
  
  isNewDrawing() {
    return this.currentDrawingId == null;
  }

  setCurrentDrawing(id) {
    this.currentDrawingId = id;
  }

  async saveOrCreate(data) {
    layerManager.saveToLocalStorage(); // TODO: When a different drawing is loaded, do this too
    if (this.isNewDrawing()) {
      bodyComponent.createNewDrawingComponent.show();
      return;
    }
    await this.save(this.currentDrawingId, response => this.saveCallback(response));
  }

  async save(id, callback) {
    let response = await patchRequest(
      "/api/drawings/mutable/" + id,
      {"data": layerManager.encodeAll()}
    );
    callback(response);
  }

  async create(data) {
    data = {...data, "data": layerManager.encodeAll()};
    return await postRequest("/api/drawings/mutable", data);
  }
  
  saveCallback(response) {
    if (response.error.length > 0) {
      bodyComponent.informerComponent.report(response.error, "bad");
      return;
    }
    bodyComponent.informerComponent.report("Successfully saved!", "good");
  }

  createCallback(response) {
    if (response.id > -1) {
      bodyComponent.informerComponent.report("Successfully saved!", "good");
      this.setCurrentDrawing(response.id);  
      bodyComponent.createNewDrawingComponent.hide();
      return;
    }
    if (response.error.length > 0) {
      bodyComponent.informerComponent.report(response.error, "bad");
      bodyComponent.createNewDrawingComponent.hide();
    }
  }
}

class CreateNewDrawingComponent extends PopupComponent {
  css_width           = "300px";
  css_height          = "200px";
  css_marginLeft      = "calc(50vw - 150px)";

  disableModes = true;

  defineChildren() {
    return [
      new Component({
        css_width: "100%",
        css_height: "15%",
        css_textAlign: "center",
        value: "<h2>Save drawing</h2>",
      }),
      new FormComponent({
        accessibleBy: "formComponent",
        formFields: {"name": "Name"},
        formOnSubmit: async (data) => await drawingManager.create(data),
        formCallback: (response) => drawingManager.createCallback(response),
        formSubmitValue: "Save!",
        formFieldProps: {
          css_width: "100%",
          css_marginBottom: "6px",
        },
        css_width: "100%",
      }),
      new Component({
        css_width: "100%",
        css_marginTop: "20px",
        value: "Please give your drawing a name.",
        css_textAlign: "center",
      }),
    ]
  }

}

class InputComponent extends Component {
  type = "input"

  css_height           = "40px";
  css_border           = "1px solid";
  css_borderRadius     = "10px";
  css_fontFamily       = "bodyFont";
  css_outline          = "none";

  defineTheme() {
    this.css("backgroundColor", "buttonBgColor");
    this.css("borderColor", "buttonBorderColor");
    this.css("color", "buttonFgColor");
  }
}

class FormComponent extends Component {
  // Example usage...
  // formFields = {"name", "Display Name", ...}
  // formFieldProps = {css_X: "X", ...}
  // formOnSubmit = (data) => something(data) 
  // formSubmitValue = "Sign Up"
  //
  // All fields can be accessed on this form object by formField<Fieldname>

  defineChildren() {
    this.formFieldComponents = {};
    for (let field in this.formFields) {
      let displayName = this.formFields[field];
      let inputType = displayName.toLowerCase().includes("pass")? "password": "text";
      this.formFieldComponents[field] = new InputComponent({
        accessibleBy: "formField" + field.substring(0, 1).toUpperCase() + field.substring(1),
        prop_type: inputType,
        prop_placeholder: displayName,
        ...this.formFieldProps,
      });
    }
    this.formSubmitButtonComponent = new ButtonComponent({
      value: this.formSubmitValue,
      css_margin: "0 auto",
      css_display: "block",
      css_width: "30%",
      on_click: async () => this.formSubmit()
    });
    return [...Object.values(this.formFieldComponents), this.formSubmitButtonComponent];
  }

  formClear() {
    for (let field in this.formFieldComponents) {
      this.formFieldComponents[field].setValue("");
    }
  }

  async formSubmit() {
    let data = {};
    for (let field in this.formFieldComponents) {
      data[field] = this.formFieldComponents[field].getValue();
      if (!data[field].length) return;
    }
    let response = await this.formOnSubmit(data);
    this.formCallback(response);
  }
}

class UserSignUpComponent extends PopupComponent {
  css_width           = "300px";
  css_height          = "250px";
  css_marginLeft      = "calc(50vw - 150px)";

  disableModes = true;

  defineChildren() {
    return [
      new Component({
        css_width: "100%",
        css_height: "15%",
        css_textAlign: "center",
        value: "<h2>Sign Up</h2>",
      }),
      new FormComponent({
        accessibleBy: "formComponent",
        formFields: {"email": "Email", "password": "Password"},
        formOnSubmit: (data) => userManager.signup(data),
        formCallback: (response) => userManager.signupCallback(response),
        formSubmitValue: "Sign Up",
        formFieldProps: {
          css_width: "100%",
          css_marginBottom: "6px",
        },
        css_width: "100%",
      }),
      new Component({
        css_width: "100%",
        css_marginTop: "10px",
        css_textAlign: "center",
        value: "Sign up to save and manage drawings. Your email is only used for account recovery.",
      }),
    ]
  }

}

class UserLoginComponent extends PopupComponent {
  css_width           = "300px";
  css_height          = "250px";
  css_marginLeft      = "calc(50vw - 150px)";

  disableModes = true;

  defineChildren() {
    return [
      new Component({
        css_width: "100%",
        css_height: "15%",
        css_textAlign: "center",
        value: "<h2>Login</h2>",
      }),
      new FormComponent({
        accessibleBy: "formComponent",
        formFields: {"email": "Email", "password": "Password"},
        formOnSubmit: (data) => userManager.login(data),
        formCallback: async (response) => await userManager.loginCallback(response),
        formSubmitValue: "Login",
        formFieldProps: {
          css_width: "100%",
          css_marginBottom: "6px",
        },
        css_width: "100%",
      }),
      new Component({
        css_width: "100%",
        css_marginTop: "15px",
        css_textAlign: "center",
        value: "Forgot Password? Please visit the Help page.",
      }),
    ]
  }

}

class RightMenuComponent extends MenuComponent {
  css_height    = "100vh";
  css_width     = "130px";
  css_float     = "right";
  css_position  = "relative";
  css_zIndex    = "100";
  css_marginTop = "120px";

  buttons = [
    new MenuButtonComponent(
      {
        value: "Login",
        accessibleBy: "loginButtonComponent",
        on_click: () => bodyComponent.loginComponent.toggle(),
        css_width: "100%",
      }
    ), 
    new MenuButtonComponent(
      {
        value: "Sign Up",
        accessibleBy: "signupButtonComponent",
        on_click: () => bodyComponent.signupComponent.toggle(),
        css_width: "100%",
      }
    ), 
    new Component(
      {
        value: "",
        accessibleBy: "welcomeMsgComponent",
        css_width: "100%",
        Css_border: "none",
        css_boxShadow: "none",
        css_fontWeight: "bold",
        css_textAlign: "center",
        css_textShadow: "0px 0px 5px grey",
      }
    ), 
    new MenuButtonComponent(
      {
        value: "Logout",
        accessibleBy: "logoutButtonComponent",
        on_click: () => userManager.logout(),
        css_width: "100%",
        Css_marginTop: "20px",
      }
    ), 
    new MenuButtonComponent(
      {
        value: "Save",
        accessibleBy: "saveButtonComponent",
        on_click: async () => await drawingManager.saveOrCreate(),
        css_width: "100%",
        Css_marginTop: "40px",
      }
    ), 
    new MenuButtonComponent(
      {
        value: "My Drawings",
        accessibleBy: "myDrawingsButtonComponent",
        on_click: () => bodyComponent.signupComponent.toggle(),
        css_width: "100%",
      }
    ), 
  ]
}

async function mainServerClient() {
  bodyComponent.addChild(new UserSignUpComponent({accessibleBy: "signupComponent"}));
  bodyComponent.addChild(new UserLoginComponent({accessibleBy: "loginComponent"}));
  bodyComponent.addChild(new RightMenuComponent({accessibleBy: "rightMenuComponent"}));
  bodyComponent.addChild(new CreateNewDrawingComponent({accessibleBy: "createNewDrawingComponent"}));

  userManager = new UserManager();
  drawingManager = new DrawingManager();

  userManager.hideAll();
  await userManager.update();
}

mainServerClient(); // :)
