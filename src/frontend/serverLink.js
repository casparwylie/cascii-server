async function postRequest(url, data) {
  let response = await fetch(url,
    {
      method: 'POST',
      headers: {'Content-Type': 'application/json'},
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

class InputComponent extends Component {
  type = "input"

  css_height           = "40px";
  css_border           = "1px solid";
  css_borderRadius     = "10px";
  css_fontFamily       = "bodyFont";
  css_outline          = "none";

  defineTheme() {
    this.css("backgroundColor", "menuButtonBgColor"); // TODO: Rename theme styles, check all uses
    this.css("borderColor", "menuButtonBorderColor");
    this.css("color", "menuButtonFgColor");
  }
}

class FormComponent extends Component {
  // Example usage...
  // formFields = {"name", "Display Name", ...}
  // formFieldProps = {css_X: "X", ...}
  // formApiEndpoint = "/api/users"
  // formSubmitValue = "Sign Up"

  defineChildren() {
    this.formFieldComponents = {};
    for (let field in this.formFields) {
      let displayName = this.formFields[field];
      let inputType = displayName.toLowerCase().includes("pass")? "password": "text";
      this.formFieldComponents[field] = new InputComponent(
        {prop_type: inputType, prop_placeholder: displayName, ...this.formFieldProps}
      );
    }
    this.formSubmitButtonComponent = new ButtonComponent(
      {value: this.formSubmitValue, on_click: async () => this.formSubmit()}
    );
    return [...Object.values(this.formFieldComponents), this.formSubmitButtonComponent];
  }

  formClear() {
    for (let field in this.formFieldComponents) {
      console.log("clearing")
      this.formFieldComponents[field].setValue("");
    }
  }

  async formSubmit() {
    let data = {};
    for (let field in this.formFieldComponents) {
      data[field] = this.formFieldComponents[field].getValue();
      if (!data[field].length) return;
    }
    let response = await postRequest(this.formApiEndpoint, data);
    this.formCallback(response);
  }
}

class UserSignUpComponent extends PopupComponent {
  css_width           = "300px";
  css_height          = "250px";
  css_marginLeft      = "calc(50vw - 150px)";

  defineChildren() {
    this.form = new FormComponent({
      css_height: "100px",
      css_width: "100%",
      formFields: {"email": "Email", "password": "Password"},
      formApiEndpoint: "/api/user/",
      formSubmitValue: "Sign Up",
      formFieldProps: {
        css_width: "100%",
        css_marginBottom: "4px",
      },
      formCallback: (response) => this.submitCallback(response),
    });
    return [
      new Component({
        css_width: "100%",
        css_height: "15%",
        css_textAlign: "center",
        value: "<h2>Sign Up</h2>",
      }),
      this.form,
    ]
  }

  submitCallback(response) {
    if (response.error == "") {
      bodyComponent.informer.report("Successfully signed up!", "good");
      this.hide();
      this.form.formClear();
    } else {
      bodyComponent.informer.report(response.error, "bad");
    }
  }
}

class UserLoginComponent extends PopupComponent {
  css_width           = "300px";
  css_height          = "250px";
  css_marginLeft      = "calc(50vw - 150px)";

  defineChildren() {
    this.form = new FormComponent({
      css_height: "100px",
      css_width: "100%",
      formFields: {"email": "Email", "password": "Password"},
      formApiEndpoint: "/api/user/auth",
      formSubmitValue: "Login",
      formFieldProps: {
        css_width: "100%",
        css_marginBottom: "4px",
      },
      formCallback: (response) => this.submitCallback(response),
    });
    return [
      new Component({
        css_width: "100%",
        css_height: "15%",
        css_textAlign: "center",
        value: "<h2>Login</h2>",
      }),
      this.form,
    ]
  }

  submitCallback(response) {
    if (response.error == "") {
      bodyComponent.informer.report("Successfully logged in!", "good");
      this.hide();
      this.form.formClear();
    } else {
      bodyComponent.informer.report(response.error, "bad");
    }
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
    new MenuButtonLeftComponent("", "Login", // Rename / make new button for this menu
      [], [modeMaster.reset], [], () => bodyComponent.login.toggle()),
    new MenuButtonLeftComponent("", "Sign Up",
      [], [modeMaster.reset], [], () => bodyComponent.signUp.toggle()),
  ]
}

function mainServerClient() {
  bodyComponent.addChild(new UserSignUpComponent(), "signUp");
  bodyComponent.addChild(new UserLoginComponent(), "login");
  bodyComponent.addChild(new RightMenuComponent());
}

mainServerClient(); // :)
