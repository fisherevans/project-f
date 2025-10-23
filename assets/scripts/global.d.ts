// AUTO-GENERATED - DO NOT EDIT
// Generated at 2025-10-23T00:10:25-04:00 by go generate
// Source: internal/game/events/handler.go, internal/game/events/effect.go

// ============================================================================
// LOGGING
// ============================================================================

interface Logger {
  Debug(message: string): void;
  Info(message: string): void;
  Warn(message: string): void;
  Error(message: string): void;
}

declare const log: Logger;

// ============================================================================
// CONTEXT
// ============================================================================

interface EntityContext {
  Id(): string;
}

interface WorldStateReader {
  // Add world state methods as needed
}

// ============================================================================
// EVENT TYPES
// ============================================================================

interface EventOnInteract {
  targetId: string;
}

interface EventDialogueComplete {
  dialogueId: string;
}

interface EventChatterComplete {
  chatterId: string;
  entityId: string;
}

interface EventTimerComplete {
  createdBy: string;
  timerId: string;
  durationSeconds: number;
  triggerCount: number;
}

interface EventEntityZoneActivity {
  entityId: string;
  zoneId: string;
  isEntering: boolean;
}

interface EventWorldStateUpdated {
  key: string;
  newValue: any;
  oldValue: any;
  setBy: string;
}

interface EventCombatComplete {
  combatId: string;
  result: any;
}

// ============================================================================
// EFFECT TYPES
// ============================================================================

interface DynamicAnimationReference {
  tilesheet: string;
  name: string;
}

interface LightConfig {
  color: string;
  size: number;
  modifier: string;
}

interface EffectFunction {
  fn: any;
}

interface EffectDialogue {
  dialogueId: string;
  text: string;
}

interface EffectChatter {
  chatterId: string;
  entityId: string;
  durationSeconds: number;
  message: string;
}

interface EffectYieldElythium {
  amount: number;
}

interface EffectTimer {
  timerId: string;
  durationSeconds: number;
}

interface EffectMutateEntity {
  entityId: string;
  mode: string | undefined;
  isPassable: boolean | undefined;
  dynamicAnimations: Record<string, DynamicAnimationReference[]>;
  dynamicLights: Record<string, LightConfig[]>;
}

interface EffectSetWorldState {
  key: string;
  value: any;
}

interface EffectSetEntityLocation {
  entityId: string;
  toReference: string | undefined;
  toLocation: any | undefined;
  toEntityId: string | undefined;
}

interface EffectTeleportPlayer {
  toReference: string | undefined;
  toLocation: any | undefined;
  toEntityId: string | undefined;
  exitDirection: any | undefined;
  transitionStyle: string | undefined;
}

interface EffectPlan {
  planId: string;
  steps: any[];
}

interface EffectBlockInput {
  blocked: boolean;
}

interface EffectFade {
  fadeId: string;
  durationSeconds: number;
  autoDeactivate: boolean | undefined;
  fromColor: string | undefined;
  toColor: string | undefined;
  transitions: number;
}

interface EffectDeactivateFade {
  fadeId: string;
}

interface EffectTriggerMovement {
  entityId: string;
  direction: any;
}

interface EffectSetFollowCamera {
  entityId: string | undefined;
  resetPosition: boolean;
}

interface EffectMutateNPC {
  entityId: string;
  talkingAtEntityId: string | undefined;
  isTalking: boolean | undefined;
}

interface EffectTriggerCombat {
  combatId: string;
  opponent: any | undefined;
  background: string;
}

interface Effect {
  function?: EffectFunction;
  dialogue?: EffectDialogue;
  chatter?: EffectChatter;
  yieldElythium?: EffectYieldElythium;
  timer?: EffectTimer;
  mutateEntity?: EffectMutateEntity;
  setWorldState?: EffectSetWorldState;
  setEntityLocation?: EffectSetEntityLocation;
  teleportPlayer?: EffectTeleportPlayer;
  plan?: EffectPlan;
  blockInput?: EffectBlockInput;
  fade?: EffectFade;
  deactivateFade?: EffectDeactivateFade;
  triggerMovement?: EffectTriggerMovement;
  setFollowCamera?: EffectSetFollowCamera;
  mutateNPC?: EffectMutateNPC;
  triggerCombat?: EffectTriggerCombat;
}

interface HandlerOutput {
  state?: any;
  effects?: Effect[];
}

// ============================================================================
// ENTITY HANDLER
// ============================================================================

interface EntityHandler {
  Init?: (self: EntityContext, world: WorldStateReader, state: any) => HandlerOutput | null;
  OnInteract?: (self: EntityContext, world: WorldStateReader, state: any, event: EventOnInteract) => HandlerOutput | null;
  OnDialogueComplete?: (self: EntityContext, world: WorldStateReader, state: any, event: EventDialogueComplete) => HandlerOutput | null;
  OnChatterComplete?: (self: EntityContext, world: WorldStateReader, state: any, event: EventChatterComplete) => HandlerOutput | null;
  OnTimerComplete?: (self: EntityContext, world: WorldStateReader, state: any, event: EventTimerComplete) => HandlerOutput | null;
  OnEntityZoneActivity?: (self: EntityContext, world: WorldStateReader, state: any, event: EventEntityZoneActivity) => HandlerOutput | null;
  OnWorldStateUpdated?: (self: EntityContext, world: WorldStateReader, state: any, event: EventWorldStateUpdated) => HandlerOutput | null;
  OnCombatComplete?: (self: EntityContext, world: WorldStateReader, state: any, event: EventCombatComplete) => HandlerOutput | null;
}

declare const handler: EntityHandler;
