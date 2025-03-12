var userManager;
var drawingManager;
var routeManager;

async function postRequest(url, data) {
  bodyComponent.informerComponent.loading();
  let response = await fetch(url, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(data),
  });
  var result = await response.json();
  result.statusCode = response.status;
  bodyComponent.informerComponent.loadingFinish();
  return result;
}

async function patchRequest(url, data) {
  bodyComponent.informerComponent.loading();
  let response = await fetch(url, {
    method: "PATCH",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(data),
  });
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
  return json;
}

async function deleteRequest(url) {
  bodyComponent.informerComponent.loading();
  let result = await fetch(url, { method: "DELETE" });
  let json = await result.json();
  json.statusCode = result.status;
  bodyComponent.informerComponent.loadingFinish();
  return json;
}

function handleResponse(response, msg = "", silentErr = false) {
  if (response.error && response.error.length > 0) {
    if (!silentErr) bodyComponent.informerComponent.report(response.error, "bad");
    return false;
  }
  if (msg.length > 0) bodyComponent.informerComponent.report(msg, "good");
  return true;
}

class ServeExternalHookManager extends BaseExternalHookManager {
  async getShortKeyUrl() {
    var host = window.location.host;
    let response = await drawingManager.createImmutableDrawing();
    if (handleResponse(response, "", true)) {
      return `${window.location.protocol}//${host}/${response.short_key}`;
    }
    return "";
  }

  triggerDrawingChanged() {
    // If there is nothing, then don't offer save option.
    if (layerManager.layers.length == 0) {
      drawingManager.setSaved();
      return;
    }
    // Otherwise, there are legit changes to save.
    drawingManager.setUnsaved();
  }
}

class RouteManager {
  constructor() {
    this.routes = [];
  }

  addRoutes(...routes) {
    this.routes = this.routes.concat(routes);
  }

  handle() {
    let path = window.location.pathname.replace(/[\/]*$/, "");
    for (let [regex, func] of this.routes) {
      let result = regex.exec(path);
      if (result != null) {
        func(result.groups);
        return;
      }
    }
  }
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
    ];
  }
  guestOnlyComponents() {
    return [
      bodyComponent.rightMenuComponent.loginButtonComponent,
      bodyComponent.rightMenuComponent.signupButtonComponent,
    ];
  }

  renderLogin() {
    this.userOnlyComponents().forEach((component) => component.show());
    this.guestOnlyComponents().forEach((component) => component.hide());
    let username = this.getUsername();
    bodyComponent.rightMenuComponent.welcomeMsgComponent.setValue("Welcome, " + username);
  }

  renderLogout() {
    this.userOnlyComponents().forEach((component) => component.hide());
    this.guestOnlyComponents().forEach((component) => component.show());
  }

  hideAll() {
    // This is used so that components don't briefly appear before API calls
    // on page load.
    this.userOnlyComponents().forEach((component) => component.hide());
    this.guestOnlyComponents().forEach((component) => component.hide());
  }

  async signup(data) {
    if (handleResponse(await this.signupUser(data), "Successfully signed up!")) {
      bodyComponent.signupComponent.hide();
      bodyComponent.loginComponent.formComponent.formFieldEmail.setValue(
        bodyComponent.signupComponent.formComponent.formFieldEmail.getValue()
      );
      bodyComponent.loginComponent.formComponent.formFieldPassword.setValue(
        bodyComponent.signupComponent.formComponent.formFieldPassword.getValue()
      );
      bodyComponent.signupComponent.formComponent.formClear();
    }
  }

  async update() {
    this.user = null;
    let response = await this.getUser();
    if (handleResponse(response, "", true)) {
      this.user = { email: response.email, userId: response.id };
    }
    this.isLoggedin() ? this.renderLogin() : this.renderLogout();
  }

  async login(data) {
    if (handleResponse(await this.loginUser(data))) {
      await userManager.update();
      if (!this.isLoggedin()) return;

      // In case a different user on same browser comes along
      drawingManager.unsetCurrentDrawing();
      drawingManager.setUnsaved();

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
  update() {
    this.isSaved() ? this.setSaved() : this.setUnsaved();
  }

  isSaved() {
    return Boolean(localStorage.getItem("serverSaved"));
  }

  setUnsaved() {
    if (!userManager.isLoggedin()) return;
    localStorage.setItem("serverSaved", "");
    bodyComponent.rightMenuComponent.saveButtonComponent.show();
  }

  setSaved() {
    if (!userManager.isLoggedin()) return;
    localStorage.setItem("serverSaved", "true");
    bodyComponent.rightMenuComponent.saveButtonComponent.hide();
  }

  async ensureSave() {
    if (!this.isSaved() && window.confirm("Would you like to save your current work?")) {
      await this.saveButtonUpdateOrCreate();
      return true;
    }
    return false;
  }

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

  async startNewDrawing() {
    if (await this.ensureSave()) return;
    this.unsetCurrentDrawing();
    bodyComponent.hidePopups();
    // emptyEvent also triggers change which knows a blank canvas
    // can't be saved (until an actual change comes in).
    layerManager.refresh(() => layerManager.emptyEvent());
    bodyComponent.informerComponent.report("Blank canvas is ready!");
  }

  async saveButtonUpdateOrCreate(data) {
    // If you "Save" and the drawing is new, then open the form with the intent to create.
    // Otherwise, update the drawing data.
    layerManager.saveToLocalStorage();
    if (this.isNewDrawing()) {
      bodyComponent.editDrawingMetaComponent.showAsCreator();
      return;
    }
    if (handleResponse(await this.saveDrawing(this.getCurrentDrawing()), "Successfully saved!")) {
      this.setSaved();
    }
  }

  async editMetaFormCreate(data) {
    let response = await this.createDrawing(data);
    if (handleResponse(response)) {
      this.setCurrentDrawing(response.id);
      bodyComponent.editDrawingMetaComponent.hide();
      await bodyComponent.listDrawingsComponent.show();
      this.setSaved();
      // This get's defered so that listing loading doesn't replace it.
      bodyComponent.informerComponent.report("Successfully saved!", "good");
    }
  }

  async editMetaFormUpdate(data, drawingId) {
    let response = await this.updateMetadataDrawing(drawingId, data);
    if (handleResponse(response)) {
      await bodyComponent.listDrawingsComponent.show();
      // This get's defered so that listing loading doesn't replace it.
      bodyComponent.informerComponent.report("Successfully updated!", "good");
    }
  }

  async open(drawingId) {
    if (await this.ensureSave()) return;
    let response = await this.getDrawing(drawingId);
    if (handleResponse(response, "Successfully loaded!")) {
      layerManager.import(response.data);
      bodyComponent.hidePopups();
      // The loaded drawing is saved until a change happens
      this.setCurrentDrawing(drawingId);
      this.setSaved();
    }
  }

  async duplicate(drawingId) {
    if (await this.ensureSave()) return;
    // Same as this.open except we don't set the drawing as current so it is saved
    // as a new one.
    let response = await this.getDrawing(drawingId);
    if (handleResponse(response, "Successfully made duplicate. Saving this will create a new drawing.")) {
      layerManager.import(response.data);
      bodyComponent.hidePopups();

      this.unsetCurrentDrawing();
      // We setUnsaved because a user should be able to save a copy regardless of change.
      this.setUnsaved();
    }
  }

  async openFromShortKey(shortKey) {
    let response = await this.getImmutableDrawing(shortKey);
    if (handleResponse(response, "Successfully loaded. This is your own version of the original to edit freely.")) {
      // It's important we load into localStorage too immediately as a side effect.
      // layerManager.import currently does this impliclity with redraw.
      layerManager.import(response.data);
      bodyComponent.hidePopups();

      this.unsetCurrentDrawing();

      // A newly opened short key should always be saveable...
      this.setUnsaved();

      // This avoids a page reload switching back to the first version, given
      // a user could have edited it (more up to date in localStorage now).
      window.history.replaceState(null, document.title, "/");
    }
  }

  async delete(drawingId, callback) {
    if (handleResponse(await this.deleteDrawing(drawingId))) {
      // We do this so that saving doesn't attempt to save to a non-existent drawing.
      // We don't setUnsaved as the user wouldn't want to save a just deleted drawing
      // unless they change it (save option will appear).
      if (drawingId == this.getCurrentDrawing()) this.unsetCurrentDrawing();
      await callback();
      // This get's defered so that listing loading doesn't replace it.
      bodyComponent.informerComponent.report("Successfully deleted!", "good");
    }
  }

  async getDrawings() {
    return (await getRequest("/api/drawings/mutables")).results || [];
  }

  async getDrawing(drawingId) {
    return await getRequest("/api/drawings/mutable/" + drawingId);
  }

  async getImmutableDrawing(shortKey) {
    return await getRequest("/api/drawings/immutable/" + shortKey);
  }

  async deleteDrawing(drawingId) {
    return await deleteRequest("/api/drawings/mutable/" + drawingId);
  }

  async saveDrawing(id) {
    let data = { data: layerManager.encodeAll() };
    return await patchRequest("/api/drawings/mutable/" + id, data);
  }

  async updateMetadataDrawing(id, data) {
    return await patchRequest("/api/drawings/mutable/" + id, data);
  }

  async createDrawing(data) {
    data = { ...data, data: layerManager.encodeAll() };
    return await postRequest("/api/drawings/mutable", data);
  }

  async createImmutableDrawing() {
    let data = { data: layerManager.encodeAll() };
    return await postRequest("/api/drawings/immutable", data);
  }
}

class ListDrawingsComponent extends PopupComponent {
  css_width = "400px";
  css_marginLeft = "calc(50vw - 200px)";
  css_overflow = "auto";
  css_maxHeight = "600px";

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
    ];
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
      this.resultsComponent.addChild(
        new Component({
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
                new Component({ value: drawing.created_at, css_fontSize: "11px" }),
              ],
            }),
            new Component({
              css_display: "flex",
              css_justifyContent: "space-around",
              css_columnGap: "3px",
              children: [
                new ButtonComponent({
                  value: "Duplicate",
                  Css_height: "30px",
                  css_padding: "2px",
                  on_mousedown: () => drawingManager.duplicate(drawing.id),
                }),
                new ButtonComponent({
                  value: "Rename",
                  Css_height: "30px",
                  css_padding: "2px",
                  on_mousedown: () => bodyComponent.editDrawingMetaComponent.showAsUpdater(drawing.id, drawing),
                }),
                new ButtonComponent({
                  value: "Delete",
                  Css_height: "30px",
                  css_padding: "2px",
                  Css_color: "warningRed",
                  on_mousedown: () => drawingManager.delete(drawing.id, async () => await this.populate()),
                }),
              ],
            }),
          ],
        })
      );
    }
  }

  async show() {
    super.show();
    await this.populate();
  }
}

class EditDrawingMetaComponent extends PopupComponent {
  css_width = "300px";
  css_height = "200px";
  css_marginLeft = "calc(50vw - 150px)";

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
        formFields: { name: "Name" },
        formOnSubmit: (data) => this.createOrUpdate(data),
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
    ];
  }

  hide() {
    super.hide();
    this.formComponent.formClear();
  }

  showAsUpdater(drawingId, currentData) {
    this.forUpdating = drawingId;
    this.show();
    this.formComponent.formFieldName.setValue(currentData.name);
  }

  showAsCreator() {
    this.forUpdating = -1;
    this.show();
  }

  createOrUpdate(data) {
    if (this.forUpdating == -1) {
      drawingManager.editMetaFormCreate(data);
    } else {
      drawingManager.editMetaFormUpdate(data, this.forUpdating);
    }
  }
}

class InputComponent extends Component {
  type = "input";

  css_height = "40px";
  css_border = "1px solid";
  css_borderRadius = "10px";
  css_fontFamily = "bodyFont";
  css_outline = "none";

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
      let inputType = displayName.toLowerCase().includes("pass") ? "password" : "text";
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
      on_click: async () => this.formSubmit(),
    });
    this.listenForEnter();
    return [...Object.values(this.formFieldComponents), this.formSubmitButtonComponent];
  }

  listenForEnter() {
    document.addEventListener("keypress", (event) => {
      if (event.key == "Enter" && this.parent.visible) {
        event.preventDefault();
        this.formSubmit();
      }
    });
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
  css_width = "300px";
  css_height = "250px";
  css_marginLeft = "calc(50vw - 150px)";

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
        formFields: { email: "Email", password: "Password" },
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
    ];
  }
}

class UserLoginComponent extends PopupComponent {
  css_width = "300px";
  css_height = "250px";
  css_marginLeft = "calc(50vw - 150px)";

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
        formFields: { email: "Email", password: "Password" },
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
    ];
  }
}

class RightMenuComponent extends MenuComponent {
  css_height = "100vh";
  css_width = "130px";
  css_float = "right";
  css_position = "relative";
  css_zIndex = "100";
  css_marginTop = "120px";

  buttons = [
    new MenuButtonComponent({
      value: "Login",
      accessibleBy: "loginButtonComponent",
      on_click: () => bodyComponent.loginComponent.toggle(),
      css_width: "100%",
    }),
    new MenuButtonComponent({
      value: "Sign Up",
      accessibleBy: "signupButtonComponent",
      on_click: () => bodyComponent.signupComponent.toggle(),
      css_width: "100%",
    }),
    new Component({
      value: "",
      accessibleBy: "welcomeMsgComponent",
      css_width: "100%",
      Css_border: "none",
      css_boxShadow: "none",
      css_fontWeight: "bold",
      css_textAlign: "center",
      css_textShadow: "0px 0px 5px grey",
    }),
    new MenuButtonComponent({
      value: "Logout",
      accessibleBy: "logoutButtonComponent",
      on_click: () => userManager.logout(),
      css_width: "100%",
      Css_marginTop: "20px",
    }),
    new MenuButtonComponent({
      value: "<b>Save</b>",
      accessibleBy: "saveButtonComponent",
      on_click: async () => await drawingManager.saveButtonUpdateOrCreate(),
      css_width: "100%",
      Css_marginTop: "40px",
      Css_color: "warningRed",
    }),
    new MenuButtonComponent({
      value: "+ New",
      accessibleBy: "newButtonComponent",
      on_click: async () => drawingManager.startNewDrawing(),
      css_width: "100%",
      Css_marginTop: "20px",
    }),
    new MenuButtonComponent({
      value: "My Drawings",
      accessibleBy: "myDrawingsButtonComponent",
      on_click: async () => await bodyComponent.listDrawingsComponent.toggle(),
      css_width: "100%",
    }),
  ];
}

async function mainServerClient() {
  bodyComponent.addChild(new UserSignUpComponent({ accessibleBy: "signupComponent" }));
  bodyComponent.addChild(new UserLoginComponent({ accessibleBy: "loginComponent" }));
  bodyComponent.addChild(new RightMenuComponent({ accessibleBy: "rightMenuComponent" }));
  bodyComponent.addChild(new EditDrawingMetaComponent({ accessibleBy: "editDrawingMetaComponent" }));
  bodyComponent.addChild(new ListDrawingsComponent({ accessibleBy: "listDrawingsComponent" }));

  userManager = new UserManager();
  drawingManager = new DrawingManager();
  externalHookManager = new ServeExternalHookManager();
  routeManager = new RouteManager();

  userManager.hideAll();

  // Render user related UI
  await userManager.update();

  // Render drawing related UI
  drawingManager.update();

  routeManager.addRoutes([/^\/(?<shortkey>[\w]+)$/, (vars) => drawingManager.openFromShortKey(vars.shortkey)]);
  routeManager.handle();
}

window.addEventListener("casciiLoaded", mainServerClient); // :)
