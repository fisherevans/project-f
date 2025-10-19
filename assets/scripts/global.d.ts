// TypeScript definitions for Project F scripting API

// ============================================================================
// WORLD API
// ============================================================================

interface World {
  /** Get a world variable value */
  getVar(key: string): any;
  
  /** Check if a world variable exists */
  hasVar(key: string): boolean;
}

declare const world: World;

// ============================================================================
// ENTITY SELF
// ============================================================================

interface EntitySelf {
  /** The entity's unique ID */
  id: string;
}

// ============================================================================
// EVENTS
// ============================================================================

interface BaseEvent {
  type: string;
  entityId: string;
  timestamp: number;
}

interface InteractEvent extends BaseEvent {
  type: "Interact";
  sourceEntityId: string;
}

interface EnterZoneEvent extends BaseEvent {
  type: "EnterZone";
  zoneName: string;
}

interface LeaveZoneEvent extends BaseEvent {
  type: "LeaveZone";
  zoneName: string;
}

interface TimerEvent extends BaseEvent {
  type: "Timer";
  timerName: string;
}

interface TriggerEvent extends BaseEvent {
  type: "Trigger";
  triggerName: string;
  data?: Record<string, any>;
}

interface FlagChangedEvent extends BaseEvent {
  type: "FlagChanged";
  varName: string;
  oldValue: any;
  newValue: any;
}

interface StepEvent extends BaseEvent {
  type: "Step";
  deltaTime: number;
}

type GameEvent = 
  | InteractEvent 
  | EnterZoneEvent 
  | LeaveZoneEvent 
  | TimerEvent 
  | TriggerEvent 
  | FlagChangedEvent 
  | StepEvent;

// ============================================================================
// EFFECTS
// ============================================================================

interface SetVarEffect {
  type: "SetVar";
  data: {
    key: string;
    value: any;
  };
}

interface IncVarEffect {
  type: "IncVar";
  data: {
    key: string;
    value: number;
  };
}

interface ClearVarEffect {
  type: "ClearVar";
  data: {
    key: string;
  };
}

interface MoveToEffect {
  type: "MoveTo";
  data: {
    x: number;
    y: number;
  };
}

interface OpenDoorEffect {
  type: "OpenDoor";
  data: {
    doorId: string;
  };
}

interface CloseDoorEffect {
  type: "CloseDoor";
  data: {
    doorId: string;
  };
}

interface TriggerEffect {
  type: "Trigger";
  data: {
    triggerName: string;
    targetEntity?: string;
    data?: Record<string, any>;
  };
}

interface StartBattleEffect {
  type: "StartBattle";
  data: {
    battleId: string;
  };
}

interface SetCameraEffect {
  type: "SetCamera";
  data: {
    target: string;
  };
}

interface PlayMusicEffect {
  type: "PlayMusic";
  data: {
    musicId: string;
    fadeIn?: number;
  };
}

interface PlaySoundEffect {
  type: "PlaySound";
  data: {
    soundId: string;
    volume?: number;
  };
}

interface ShowDialogEffect {
  type: "ShowDialog";
  data: {
    text: string;
    speaker?: string;
  };
}

interface StartTimerEffect {
  type: "StartTimer";
  data: {
    timerName: string;
    duration: number;
  };
}

interface StopTimerEffect {
  type: "StopTimer";
  data: {
    timerName: string;
  };
}

interface SpawnEntityEffect {
  type: "SpawnEntity";
  data: {
    entityType: string;
    x: number;
    y: number;
    properties?: Record<string, any>;
  };
}

interface RemoveEntityEffect {
  type: "RemoveEntity";
  data: {
    entityId?: string;
  };
}

interface TransactionEffect {
  type: "Transaction";
  data: {
    effects: Effect[];
  };
}

type Effect = 
  | SetVarEffect 
  | IncVarEffect 
  | ClearVarEffect
  | MoveToEffect
  | OpenDoorEffect
  | CloseDoorEffect
  | TriggerEffect
  | StartBattleEffect
  | SetCameraEffect
  | PlayMusicEffect
  | PlaySoundEffect
  | ShowDialogEffect
  | StartTimerEffect
  | StopTimerEffect
  | SpawnEntityEffect
  | RemoveEntityEffect
  | TransactionEffect;

// ============================================================================
// PLANS
// ============================================================================

interface EffectPlan {
  type: "Effect";
  data: Effect;
}

interface WaitPlan {
  type: "Wait";
  data: {
    duration: number;
  };
}

interface SeqPlan {
  type: "Seq";
  children: Plan[];
}

interface ParPlan {
  type: "Par";
  children: Plan[];
}

interface IfPlan {
  type: "If";
  data: {
    condition: boolean;
  };
  children: [Plan] | [Plan, Plan]; // then, optional else
}

interface ChoicePlan {
  type: "Choice";
  data: {
    prompt: string;
    options: string[];
  };
  children: Plan[]; // One per option
}

type Plan = EffectPlan | WaitPlan | SeqPlan | ParPlan | IfPlan | ChoicePlan;

// ============================================================================
// HANDLER RESULT
// ============================================================================

interface HandlerResult {
  /** Effects to apply immediately */
  effects?: Effect[];
  
  /** Multi-step plan to execute */
  plan?: Plan;
  
  /** Updated entity state */
  state?: Record<string, any>;
}

// ============================================================================
// SUBSCRIPTION
// ============================================================================

interface Subscription {
  /** Zone names to listen for EnterZone events */
  EnterZone?: string | string[];
  
  /** Zone names to listen for LeaveZone events */
  LeaveZone?: string | string[];
  
  /** Trigger names to listen for */
  Trigger?: string | string[];
  
  /** Variable prefixes to watch (e.g., "map.meadow.*") */
  VarPrefix?: string | string[];
  
  /** Handler priority (higher = earlier execution) */
  Priority?: number;
}

declare const SUBSCRIPTION: Subscription;

// ============================================================================
// EVENT HANDLERS
// ============================================================================

/** Called when entity is first created */
declare function OnInit(
  self: EntitySelf,
  world: World,
  state: Record<string, any>
): HandlerResult | void;

/** Called when entity is interacted with */
declare function OnInteract(
  self: EntitySelf,
  world: World,
  state: Record<string, any>,
  event: InteractEvent
): HandlerResult | void;

/** Called when entity enters a zone */
declare function OnEnterZone(
  self: EntitySelf,
  world: World,
  state: Record<string, any>,
  event: EnterZoneEvent
): HandlerResult | void;

/** Called when entity leaves a zone */
declare function OnLeaveZone(
  self: EntitySelf,
  world: World,
  state: Record<string, any>,
  event: LeaveZoneEvent
): HandlerResult | void;

/** Called when a timer expires */
declare function OnTimer(
  self: EntitySelf,
  world: World,
  state: Record<string, any>,
  event: TimerEvent
): HandlerResult | void;

/** Called when a trigger is activated */
declare function OnTrigger(
  self: EntitySelf,
  world: World,
  state: Record<string, any>,
  event: TriggerEvent
): HandlerResult | void;

/** Called when a watched variable changes */
declare function OnFlagChanged(
  self: EntitySelf,
  world: World,
  state: Record<string, any>,
  event: FlagChangedEvent
): HandlerResult | void;

/** Called periodically for background processing */
declare function OnStep(
  self: EntitySelf,
  world: World,
  state: Record<string, any>,
  event: StepEvent
): HandlerResult | void;

// ============================================================================
// UTILITY FUNCTIONS
// ============================================================================

/** Get deterministic game time (milliseconds since epoch) */
declare function getTime(): number;

/** Get deterministic random number [0, 1) */
declare function random(): number;
