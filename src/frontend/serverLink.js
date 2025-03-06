var userManager;
var drawingManager;

async function postRequest(url, data) {
  bodyComponent.informerComponent.loading();
  let response = await fetch(url,
    {
      method: "POST",
      headers: {"Content-Type": "application/json"},
      body: JSON.stringify(data),
    }
  );
  var result = await response.json();
  result.statusCode = response.status;
  bodyComponent.informerComponent.loadingFinish();
  return result;
}

async function patchRequest(url, data) {
  bodyComponent.informerComponent.loading();
  let response = await fetch(url,
    {
      method: "PATCH",
      headers: {"Content-Type": "application/json"},
      body: JSON.stringify(data),
    }
  );
  var result = await response.json();
  result.statusCode = response.status;
  bodyComponent.informerComponent.loadingFinish();
  return result;
}

async function getRequest(url) {
  bodyComponent.informerComponent.loading();
  let result = await fetch(url);
  let json = await result.json();
  json.statusCode = result.status;
  bodyComponent.informerComponent.loadingFinish();
  return json
}

async function deleteRequest(url) {
  bodyComponent.informerComponent.loading();
  let result = await fetch(url, {method: "DELETE"});
  let json = await result.json();
  json.statusCode = result.status;
  bodyComponent.informerComponent.loadingFinish();
  return json
}

function handleResponse(response, msg="", silentErr=false) {
  if (response.error && response.error.length > 0) {
    if (!silentErr) bodyComponent.informerComponent.report(response.error, "bad");
    return false;
  } 
  if (msg.length > 0) bodyComponent.informerComponent.report(msg, "good");
  return true;
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
      bodyComponent.rightMenuComponent.newButtonComponent,
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

  async signup(data) {
    if (handleResponse(await this.signupUser(data), "Successfully signed up!")) {
      bodyComponent.signupComponent.hide();
      bodyComponent.loginComponent.formComponent.formFieldEmail.setValue(
        bodyComponent.signupComponent.formComponent.formFieldEmail.getValue()
      )
      bodyComponent.loginComponent.formComponent.formFieldPassword.setValue(
        bodyComponent.signupComponent.formComponent.formFieldPassword.getValue()
      )
      bodyComponent.signupComponent.formComponent.formClear();
    }
  }

  async update() {
    this.user = null;
    let response = await this.getUser();
    if (handleResponse(response, "", true)) {
      this.user = {email: response.email, userId: response.id};
    }
    this.isLoggedin()? this.renderLogin(): this.renderLogout();
  }

  async login(data) {
    if (handleResponse(await this.loginUser(data))) {
      await userManager.update();
      if (!this.isLoggedin()) return;
      
      // In case a different user on same browser comes along
      drawingManager.unsetCurrentDrawing();

      bodyComponent.informerComponent.report("Successfully logged in!", "good");
      bodyComponent.loginComponent.hide();
      bodyComponent.loginComponent.formComponent.formClear();
    }
  }

  async logout() {
    if (handleResponse(await this.logoutUser())) {
      await this.update();
      if (this.isLoggedin()) return;
      bodyComponent.hidePopups();
      bodyComponent.informerComponent.report("Successfully logged out!", "good");
    }
  }

  async getUser() {
    return await getRequest("/api/user/");
  }

  async logoutUser() {
    return await getRequest("/api/user/logout");
  }

  async loginUser(data) {
    return await postRequest("/api/user/auth", data);
  }

  async signupUser(data) {
    return await postRequest("/api/user/", data);
  }
}


class DrawingManager {
  
  isNewDrawing() {
    return !this.getCurrentDrawing();
  }

  getCurrentDrawing() {
    return localStorage.getItem("currentDrawingId");
  }

  setCurrentDrawing(id) {
    localStorage.setItem("currentDrawingId", id);
  }

  unsetCurrentDrawing() {
    this.setCurrentDrawing("");
  }

  startNewDrawing() {
    this.unsetCurrentDrawing();
    layerManager.refresh(() => layerManager.empty());
    bodyComponent.informerComponent.report("Blank canvas is ready!");
  }

  async saveOrCreate(data) {
    layerManager.saveToLocalStorage();
    if (this.isNewDrawing()) {
      bodyComponent.createNewDrawingComponent.show();
      return;
    }
    handleResponse(await this.saveDrawing(this.getCurrentDrawing()), "Successfully saved!");
  }

  async create(data) {
    let response = await this.createDrawing(data);
    if (handleResponse(response, "Successfully saved!")) {
      this.setCurrentDrawing(response.id);  
      bodyComponent.createNewDrawingComponent.hide();
      bodyComponent.listDrawingsComponent.show();
    }
  }

  async open(drawingId) {
    let response = await this.getDrawing(drawingId);
    if (handleResponse(response)) {
      this.setCurrentDrawing(drawingId);
      // It's important we load into localStorage too immediately as a side effect.
      // layerManager.import currently does this impliclity with redraw.
      layerManager.import(response.data);
      bodyComponent.hidePopups();
    } 
  }

  async delete(drawingId, callback) {
    if (handleResponse(await this.deleteDrawing(drawingId), "Successfully deleted!")) {
      if (drawingId == this.getCurrentDrawing()) this.unsetCurrentDrawing();
      callback();
    }
  }

  async getDrawings() {
    return (await getRequest("/api/drawings/mutables")).results || [];
  }

  async getDrawing(drawingId) {
    return await getRequest("/api/drawings/mutable/" + drawingId);
  }

  async deleteDrawing(drawingId) {
    return await deleteRequest("/api/drawings/mutable/" + drawingId);
  }

  async saveDrawing(id) {
    let data = {"data": layerManager.encodeAll()};
    return  await patchRequest("/api/drawings/mutable/" + id, data);
  }

  async createDrawing(data) {
    data = {...data, "data": layerManager.encodeAll()};
    return await postRequest("/api/drawings/mutable", data);
  }

}


class ListDrawingsComponent extends PopupComponent {
  css_width           = "400px";
  css_marginLeft      = "calc(50vw - 200px)";
  css_overflow        = "auto";
  css_maxHeight       = "600px";

  disableModes = true;

  defineChildren() {
    return [
      new Component({
        accessibleBy: "headingComponent",
        css_width: "100%",
        css_height: "15%",
        css_textAlign: "center",
        value: "<h2>My Drawings</h2>",
      }),
      new Component({
        accessibleBy: "resultsComponent",
        css_width: "100%",
      }),
    ]
  }

  async populate() {
    let drawings = await drawingManager.getDrawings();
    this.resultsComponent.setValue("");
    if (drawings.length == 0) {
      this.headingComponent.setValue("<h2>No saved drawings!</h2>");
      return;
    }
    this.headingComponent.setValue("<h2>My Drawings</h2>");
    for (let drawing of drawings) {
      this.resultsComponent.addChild(new Component({
        css_padding: "5px",
        css_fontSize: "15px",
        css_borderColor: "bodyFgColor",
        css_display: "flex",
        css_justifyContent: "space-between",
        width: "100%",
        children: [
          new Component({
            children: [
              new Component({
                value: drawing.name,
                css_cursor: "pointer",
                on_mousedown: () => drawingManager.open(drawing.id),
              }),
              new Component({value: drawing.created_at, css_fontSize: "11px"}),
            ]
          }),
          new Component({
            css_display: "flex",
            css_justifyContent: "space-around",
            css_columnGap: "3px",
            children: [
              new ButtonComponent({
                value: "Export",
                Css_height: "30px",
                css_padding: "2px"
              }),
              new ButtonComponent({
                value: "Rename",
                Css_height: "30px",
                css_padding: "2px"
              }),
              new ButtonComponent({
                value: "Delete",
                Css_height: "30px",
                css_padding: "2px",
                Css_color: "warningRed",
                on_mousedown: async () => await drawingManager.delete(
                  drawing.id,
                  () => this.populate(),
                ),
              }),
            ]
          }),
        ],
      }));
    }
  }

  async show() {
    super.show();
    await this.populate();
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
        value: "<h2>Save Drawing</h2>",
      }),
      new FormComponent({
        accessibleBy: "formComponent",
        formFields: {"name": "Name"},
        formOnSubmit: (data) => drawingManager.create(data),
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
        value: "+ New",
        accessibleBy: "newButtonComponent",
        on_click: async () => drawingManager.startNewDrawing(),
        css_width: "100%",
        Css_marginTop: "40px",
      }
    ), 
    new MenuButtonComponent(
      {
        value: "Save",
        accessibleBy: "saveButtonComponent",
        on_click: async () => await drawingManager.saveOrCreate(),
        css_width: "100%",
      }
    ), 
    new MenuButtonComponent(
      {
        value: "My Drawings",
        accessibleBy: "myDrawingsButtonComponent",
        on_click: async () => await bodyComponent.listDrawingsComponent.toggle(),
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
  bodyComponent.addChild(new ListDrawingsComponent({accessibleBy: "listDrawingsComponent"}));

  userManager = new UserManager();
  drawingManager = new DrawingManager();

  userManager.hideAll();
  await userManager.update();
}

mainServerClient(); // :)
